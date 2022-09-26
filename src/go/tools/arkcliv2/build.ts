import * as ark from 'arksdk'
import * as filepath from 'arksdk/filepath'
import * as root from '../../../../build'
import * as golang from '../../../docker/golang.1.17-alpine/build'
import * as pkm from '../../tools/package-manager/build'
import config from '../../../../build.config'

const packageData = { name: 'arkcliv2' }
const serviceAccountSecret = 'ark-builder-service-account'
const baseBucketPath = `${config.cdnBucket}/assets/ark`
const baseCDNPAth = `${config.cdnURL}/assets/ark`
const dockerTemplatePath = filepath.fromRoot('src/docker/go_tools.v3.docker')
const version = JSON.parse(filepath.load('./version.json'))

const ignoredByBin = [
  'build.ts',
  'publish.docker',
  'install.sh'
]

const bin = ark.actions.buildDockerImage({
  name: 'bin',
  dependsOn: [
    golang.bin,
    root.goModules,
  ],
  attributes: {
    repo: `${config.baseRepo}/ark/bin`,
    dockerfile: filepath.loadAsTemplate(dockerTemplatePath, {
      package: packageData,
      platforms: ['darwin', 'linux'],
      from: golang.bin.attributes.url,
      modules: root.goModules.attributes.url,
    }),
    disableEntrypointInjection: true
  },
  sourceFiles: filepath.glob(
    './**/*',
    'src/go/lib/**/*',
    'src/go/tools/pack/**/*',
    'src/jsonnet/lib/**/*'
  ).filter(f => !ignoredByBin.some(i => f.endsWith(i)))
})

// const downloadBaseURL = `${baseCDNPAth}/versions/${bin.hash}`
// const installer = ark.actions.localFile({
//   name: 'generate.installer',
//   attributes: {
//     filename: 'install.sh',
//     content: filepath.loadAsTemplate('./install.sh', {
//       downloadBaseURL
//     }, false, {
//       objectLeft: "{{", objectRight: "}}",
//     }),
//   },
//   sourceFiles: ['./install.sh']
// })

const publish = ark.actions.buildDockerImage({
  name: 'publish',
  dependsOn: [bin, pkm.bin],
  attributes: {
    repo: `${config.baseRepo}/ark/publish`,
    disableEntrypointInjection: true,
    dockerfile: filepath.loadAsTemplate('./publish.docker', {
      package: packageData,
      arkcliArtifactURL: bin.attributes.url,
      pkmArtifactURL: pkm.bin.attributes.url,
      serviceAccountSecret,
      bucket: config.cdnBucket,
      prefix: 'assets/ark',
      version: version.version,
      // installScriptPath: installer.attributes.renderedFilePath
    }),
    secrets: [serviceAccountSecret]
  },
  sourceFiles: [
    './publish.docker',
    // installer.attributes.renderedFilePath
  ],

  // the rendered file won't exist until after graph execution
  // the hash of the rendered file is deterministic by its own source files
  // therefore we can safely omit this
  // ignoreFileNotExistsError: true,
  // excludeFromHash: {
  //   sourceFiles: [installer.attributes.renderedFilePath]
  // }
})