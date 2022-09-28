const files = filepath.glob("**")
const nodeBaseImage = ark.actions.buildDockerImage({
  name: 'nodeBaseImage',
  dockerfile: '',
  repo: '',
  sourceFiles: files,
})