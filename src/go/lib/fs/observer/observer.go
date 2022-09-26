package observer

import (
	"crypto/sha256"
	"fmt"
	"hash"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/myfintech/ark/src/go/lib/logz"

	"github.com/myfintech/ark/src/go/lib/pattern"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"

	"golang.org/x/net/context"

	"github.com/reactivex/rxgo/v2"

	"github.com/myfintech/ark/src/go/lib/log"

	"gopkg.in/src-d/go-git.v4/plumbing/format/gitignore"

	"github.com/myfintech/ark/src/go/lib/fs"
	"github.com/myfintech/ark/src/go/lib/utils"
	"github.com/myfintech/ark/src/go/lib/utils/cryptoutils"
	"github.com/myfintech/ark/src/go/lib/watchman"
	"github.com/myfintech/ark/src/go/lib/watchman/wexp"
)

var fields = logz.Fields{
	"system": "fs.rx.observer",
}

// Observer watches the filesystem for changes and notifies interested processes
type Observer struct {
	Root             string
	Ignore           []string
	Logger           logz.FieldLogger
	GitIgnoreMatcher gitignore.Matcher
	FileSystemStream rxgo.Observable

	fileSystemCache sync.Map
	fileMatchCache  sync.Map

	WaitForInitialScan   chan bool
	initialScanCompleted bool

	WatchmanClient *watchman.Client
}

// FileMatchCache uses an pattern.Matcher to capture and cache a list of files passing through the observers.FileSystemStream
type FileMatchCache struct {
	Matcher  *pattern.Matcher
	Files    sync.Map
	Hash     hash.Hash
	PrevHash hash.Hash
}

// Len returns the length of the Files sync.Map
func (f *FileMatchCache) Len() int {
	length := 0
	f.Files.Range(func(_, _ interface{}) bool {
		length++
		return true
	})
	return length
}

// FilesList returns the file list as a slice of fs.Files
func (f *FileMatchCache) FilesList() (fileList []*fs.File) {
	f.Files.Range(func(key, value interface{}) bool {
		if file, ok := value.(*fs.File); ok {
			fileList = append(fileList, file)
		}
		return true
	})
	return
}

// FileFilter should return true if the file should be retained
type FileFilter func(file string) bool

// FilesStringList returns the file list as a slice of fs.Files
func (f *FileMatchCache) FilesStringList(filter FileFilter) (fileList []string) {
	f.Files.Range(func(key, value interface{}) bool {
		if file, ok := value.(*fs.File); ok {
			if filter == nil || filter(file.Name) {
				fileList = append(fileList, file.Name)
			}
		}
		return true
	})
	return
}

// GzipArchive writes all cached files to a dest io.Writer
// Allows passing a prefix to trim from file names and a tar archive injector
func (f *FileMatchCache) GzipArchive(prefix string, dest io.Writer, filter FileFilter, inject fs.TarInjectorFunc) error {
	return fs.GzipTarFiles(f.FilesStringList(filter), prefix, dest, inject)
}

// Changed compares the byte values of Hash and PrevHash to detect changes
func (f *FileMatchCache) Changed() bool {
	return cryptoutils.CompareHashes(f.Hash, f.PrevHash) == false
}

// Copy returns a deep copy of this FileMatchCache
func (f FileMatchCache) Copy() *FileMatchCache {
	return &f
}

// ChangeNotification represents a change in the file system and any matches that were registered
type ChangeNotification struct {
	Files   []*fs.File
	Matched map[string]*FileMatchCache
}

// GetMatchCache returns a match cache by its registered key
func (o *Observer) GetMatchCache(key string) (*FileMatchCache, bool) {
	val, ok := o.fileMatchCache.Load(key)
	if !ok {
		return nil, ok
	}

	fmc, ok := val.(*FileMatchCache)
	return fmc, ok
}

// SortedFiles returns a sorted list of files by name
func (f *FileMatchCache) SortedFiles() []*fs.File {
	var files []*fs.File
	f.Files.Range(func(key, value interface{}) bool {
		if file, ok := value.(*fs.File); ok {
			files = append(files, file)
		}
		return true
	})

	sort.Slice(files, func(i, j int) bool {
		return files[i].RelName < files[j].RelName
	})
	return files
}

// SortedFilesStringList returns a sorted list of files by name
func (f *FileMatchCache) SortedFilesStringList() []string {
	var files []string
	f.Files.Range(func(key, value interface{}) bool {
		if file, ok := value.(*fs.File); ok {
			files = append(files, file.RelName)
		}
		return true
	})

	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})
	return files
}

// ComputeHash generates a sha256 hash of the cached files
func (f *FileMatchCache) ComputeHash() {
	f.PrevHash = f.Hash
	rootHash := sha256.New()
	for _, file := range f.SortedFiles() {
		if file.IsRegular() {
			_, err := fmt.Fprintf(rootHash, "%s %s\n", file.Hash, file.RelName)
			if err != nil {
				panic(errors.Wrap(err, "observer.FileMatchCache#ComputeHash failed to write to the rootHash"))
			}
		}
	}
	f.Hash = rootHash
}

// Reindex executes a get on the FileSystemSteam to trigger file system indexing
// this function automatically cancels observation of the file system
func (o *Observer) Reindex() (*ChangeNotification, error) {
	o.Logger.Trace("starting filesystem stream")
	ctx, cancel := context.WithCancel(context.TODO())

	defer func() {
		o.Logger.Trace("closing filesystem stream")
		cancel()
	}()

	i, err := o.FileSystemStream.First(rxgo.WithContext(ctx)).Get()
	if err != nil {
		return nil, err
	}

	if i.E != nil {
		return nil, i.E
	}

	if cn, ok := i.V.(*ChangeNotification); ok {
		return cn, nil
	}

	return nil, errors.New("observer returned nil ChangeNotification")
}

// AddFileMatcher adds a fileMatcher which is used to group files observed files
func (o *Observer) AddFileMatcher(key string, matcher *pattern.Matcher) error {
	if err := matcher.Compile(); err != nil {
		return err
	}
	o.fileMatchCache.Store(key, &FileMatchCache{
		Matcher:  matcher,
		Hash:     sha256.New(),
		PrevHash: sha256.New(),
		Files:    sync.Map{},
	})
	return nil
}

func (o *Observer) watchmanFileSystemProducer(ctx context.Context, ch chan<- rxgo.Item) {
	defer o.Logger.Trace("subscription has closed")

	o.Logger.Trace("connecting to the watchman socket")
	wm := o.WatchmanClient
	// wm, err := watchman.Connect(ctx, 10)
	// if err != nil {
	//	ch <- rxgo.Error(err)
	//	return
	// }
	//
	// defer func() {
	//	_ = wm.Close()
	// }()

	o.Logger.Tracef("setting project root %s", o.Root)
	resp, err := wm.WatchProject(watchman.WatchProjectOptions{
		Directory: o.Root,
	})

	if err != nil {
		ch <- rxgo.Error(err)
		return
	}

	if resp.RelPath != "" {
		o.Root = filepath.Join(resp.Watch, resp.RelPath)
	}

	expression := wexp.AllOf(wexp.Match("*", "basename", map[string]interface{}{
		"includedotfiles": true,
	}))

	for _, s := range o.Ignore {
		expression = append(expression, wexp.Not(wexp.Match(s, "wholename", map[string]interface{}{
			"includedotfiles": true,
		})))
	}

	o.Logger.Tracef("subscribing to %s", o.Root)
	o.Logger.Trace(utils.MarshalJSONSafe(expression, true))
	// TODO: pass subscription id or use UUID
	subID := utils.UUIDV4()
	err = wm.Send(watchman.RawPDUCommand{"subscribe", o.Root, subID, &watchman.QueryFilter{
		Fields:     watchman.BasicFields(),
		Expression: expression,
		DeferVcs:   true,
	}})

	if err != nil {
		ch <- rxgo.Error(err)
		return
	}

	go func() {
		<-ctx.Done()
		o.Logger.Trace("context was canceled")
		_ = wm.Close()
	}()

	for {
		o.Logger.Trace("waiting for notification")
		subResp, recErr := wm.Receive()

		o.Logger.Tracef("message received")
		if netOpErr, ok := recErr.(*net.OpError); ok {
			o.Logger.Tracef("a network operation error occurred %v", netOpErr)
			break
		}
		if errors.Is(recErr, io.EOF) {
			recErr = errors.Wrap(recErr, "the watchman server was shutdown")
		}
		if recErr != nil {
			o.Logger.Tracef("an error occurred on the watchman socket %v", recErr)
			ch <- rxgo.Error(recErr)
		}

		if _, ok := subResp["subscribe"]; ok {
			continue
		}

		o.Logger.Trace("decoding response")
		change := watchman.SubscriptionChangeFeedResponse{}
		recErr = subResp.Decode(&change)
		if recErr != nil {
			ch <- rxgo.Error(recErr)
			continue
		}

		var files []*fs.File
		for _, file := range change.Files {
			files = append(files, &fs.File{
				Name:          filepath.Join(o.Root, file.Name),
				Exists:        file.Exists,
				New:           file.New,
				Type:          file.Type,
				Hash:          file.ContentSHA1Hex,
				SymlinkTarget: file.SymlinkTarget,
				RelName:       file.Name,
			})
			o.Logger.WithFields(log.Fields{
				"type": file.Type,
				"name": file.Name,
				"hash": file.ContentSHA1Hex,
			}).Trace("observed")
		}

		ch <- rxgo.Of(files)
	}
}

// nativeFileWalk returns a slice of files
func (o *Observer) nativeFileWalk() (files []*fs.File, err error) {
	err = filepath.Walk(o.Root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		file, err := fs.NativeFileToSynthetic(o.Root, path, info)
		if err != nil {
			return err
		}

		if o.GitIgnoreMatcher != nil && o.GitIgnoreMatcher.Match(fs.Split(file.RelName), info.IsDir()) {
			o.Logger.Debugf("%s was ignored by git", file.Name)
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		files = append(files, file)

		o.Logger.WithFields(log.Fields{
			"type": file.Type,
			"name": file.Name,
			"hash": file.Hash,
		}).Trace("observed")

		return nil
	})
	return
}

func (o *Observer) nativeFileSystemProducer(ctx context.Context, ch chan<- rxgo.Item) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ch <- rxgo.Error(err)
		return
	}
	defer func() { _ = watcher.Close() }()

	files, err := o.nativeFileWalk()

	if err != nil {
		ch <- rxgo.Error(err)
		return
	}

	for _, file := range files {
		if file.Type == "d" {
			err = watcher.Add(file.Name)
			if err != nil {
				ch <- rxgo.Error(err)
			}
		}
	}

	// emit initial file bundle
	ch <- rxgo.Of(files)

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-watcher.Events:
			if isNativeFileDeleteOp(event) {
				value, ok := o.fileSystemCache.Load(event.Name)
				if !ok {
					continue
				}
				file := value.(*fs.File)
				if file.Type == "d" {
					if rmErr := watcher.Remove(event.Name); rmErr != nil {
						ch <- rxgo.Error(rmErr)
					}
				}
				file.Exists = false
				ch <- rxgo.Of([]*fs.File{file})
				continue
			}

			info, lstatErr := os.Lstat(event.Name)
			if lstatErr != nil {
				ch <- rxgo.Error(lstatErr)
				continue
			}
			if info.IsDir() && isNativeFileCreateOp(event) {
				watchErr := watcher.Add(event.Name)
				if watchErr != nil {
					ch <- rxgo.Error(watchErr)
				}
			}
			file, fsErr := fs.NativeFileToSynthetic(o.Root, event.Name, info)
			if fsErr != nil {
				ch <- rxgo.Error(fsErr)
				continue
			}
			ch <- rxgo.Of([]*fs.File{file})
		case watchErr := <-watcher.Errors:
			ch <- rxgo.Error(watchErr)
			continue
		}
	}
}

func (o *Observer) nativeFileSystemWalker(_ context.Context, ch chan<- rxgo.Item) {
	files, err := o.nativeFileWalk()
	if err != nil {
		ch <- rxgo.Error(err)
		return
	}

	ch <- rxgo.Of(files)
}

func (o *Observer) fileSystemIndexer(ctx context.Context, i interface{}) (interface{}, error) {
	if ctx.Err() != nil {
		return i, ctx.Err()
	}

	files := i.([]*fs.File)
	for _, file := range files {
		o.fileSystemCache.Store(file.Name, file)
	}
	return i, nil
}

func (o *Observer) fileMatchIndexer(ctx context.Context, i interface{}) (interface{}, error) {
	if ctx.Err() != nil {
		return i, ctx.Err()
	}

	files := i.([]*fs.File)
	changeNotification := &ChangeNotification{
		Files:   files,
		Matched: make(map[string]*FileMatchCache),
	}

	o.fileMatchCache.Range(func(key, value interface{}) bool {
		cache := value.(*FileMatchCache)

		for _, file := range files {
			if cache.Matcher.Check(file.Name) {
				if file.Exists {
					cache.Files.Store(file.Name, file)
				} else {
					cache.Files.Delete(file.Name)
				}
			}
		}
		cache.ComputeHash()
		if cache.Changed() {
			changeNotification.Matched[key.(string)] = cache.Copy()
		}
		return true
	})

	return changeNotification, nil
}

func (o *Observer) markInitialScanComplete(_ context.Context, i interface{}) (interface{}, error) {
	if !o.initialScanCompleted {
		o.initialScanCompleted = true
		o.WaitForInitialScan <- o.initialScanCompleted
		close(o.WaitForInitialScan)
	}
	return i, nil
}

func (o *Observer) fsProducer(nativeMode bool, watch bool) []rxgo.Producer {
	if !watch {
		return []rxgo.Producer{o.nativeFileSystemWalker}
	}
	if nativeMode {
		return []rxgo.Producer{o.nativeFileSystemProducer}
	}

	return []rxgo.Producer{o.watchmanFileSystemProducer}
}

func (o *Observer) mustNotBeGitIgnored(_ context.Context, i interface{}) (interface{}, error) {
	if o.GitIgnoreMatcher == nil {
		o.Logger.Debug("no gitignore matcher supplied")
		return i, nil
	}

	files := i.([]*fs.File)
	var filteredFiles []*fs.File

	for _, file := range files {
		if strings.HasSuffix(file.RelName, ".gitignore") {
			o.Logger.Debugf("%s is a .gitignore file, excluding from file match cache", file.Name)
			continue
		}
		if o.GitIgnoreMatcher.Match(fs.Split(file.RelName), file.IsDir()) {
			o.Logger.Debugf("%s was ignored by git", file.Name)
			continue
		}
		filteredFiles = append(filteredFiles, file)
	}

	return filteredFiles, nil
}

func (o *Observer) filterSwapFiles(_ context.Context, i interface{}) (interface{}, error) {
	files := i.([]*fs.File)
	var filteredFiles []*fs.File
	for _, file := range files {
		if !strings.HasSuffix(file.RelName, "~") &&
			!strings.HasSuffix(file.RelName, ".swp") &&
			!strings.HasSuffix(file.RelName, ".swx") {
			filteredFiles = append(filteredFiles, file)
		}
	}
	return filteredFiles, nil
}

// NewObserver returns a deferred file system observer
func NewObserver(
	nativeMode, watch bool,
	root string,
	ignore []string,
	gitIgnoreMatcher gitignore.Matcher,
	logger logz.FieldLogger,
	watchmanClient *watchman.Client,
) *Observer {
	o := &Observer{
		Root:               root,
		Ignore:             ignore,
		GitIgnoreMatcher:   gitIgnoreMatcher,
		WaitForInitialScan: make(chan bool, 1),
		Logger:             logger.Child(logz.WithFields(fields)),
		WatchmanClient:     watchmanClient,
	}
	o.FileSystemStream = rxgo.
		Defer(o.fsProducer(nativeMode, watch)).
		Map(o.mustNotBeGitIgnored).
		Map(o.filterSwapFiles).
		Map(o.fileSystemIndexer).
		Map(o.fileMatchIndexer).
		Map(o.markInitialScanComplete)
	return o
}

// NewObserverWithoutIndexing returns a deferred file system observer without the indexing capabilities
func NewObserverWithoutIndexing(
	nativeMode, watch bool,
	root string,
	ignore []string,
	gitIgnoreMatcher gitignore.Matcher,
	logger logz.FieldLogger,
	watchmanClient *watchman.Client,
) *Observer {
	o := &Observer{
		Root:               root,
		Ignore:             ignore,
		GitIgnoreMatcher:   gitIgnoreMatcher,
		WaitForInitialScan: make(chan bool, 1),
		Logger:             logger,
		WatchmanClient:     watchmanClient,
	}
	o.FileSystemStream = rxgo.
		Defer(o.fsProducer(nativeMode, watch)).
		Map(o.mustNotBeGitIgnored).
		Map(o.filterSwapFiles)

	return o
}

func isNativeFileCreateOp(event fsnotify.Event) bool {
	return event.Op&fsnotify.Create == fsnotify.Create
}

func isNativeFileDeleteOp(event fsnotify.Event) bool {
	return event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename
}
