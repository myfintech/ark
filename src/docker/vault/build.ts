import * as ark from 'arksdk'
import * as filepath from 'arksdk/filepath'

export const image = ark.actions.buildDockerImage({
  name: 'image',
  attributes: {
    repo: 'gcr.io/[insert-google-project]/ark/dev/vault',
    dockerfile: filepath.load('./Dockerfile')
  },
  sourceFiles: [
    './config.hcl',
    './entrypoint.sh',
    './unseal.json',
    ...filepath.glob("./data/**/*")
  ]
})
