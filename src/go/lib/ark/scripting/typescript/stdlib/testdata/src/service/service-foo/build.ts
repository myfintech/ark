import * as arksdk from 'arksdk'
import * as filepath from 'arksdk/filepath'
import config from '../../config'

export const image = arksdk.actions.buildDockerImage({
  name: 'image',
  sourceFiles: filepath.glob('**/*.ts'),
  attributes: {
    repo: config.repoBaseURL,
    dockerfile: filepath.loadAsTemplate('./Dockerfile', {
      envVar: 'test'
    })
  }
})

export const deploy = arksdk.actions.deploy({
  name: 'deploy',
  dependsOn: [image],
  attributes: {
    manifest: filepath.loadAsTemplate('./test_manifest.yaml', {
      name: 'application-service',
      imageURL: image.attributes.url
    })
  }
})