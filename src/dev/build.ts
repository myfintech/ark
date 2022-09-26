import * as ark from 'arksdk'
import microservice from 'ark/plugins/@mantl/sre/microservice'

const image = ark.actions.buildDockerImage({
  name: 'image',
  attributes: {
    repo: 'gcr.io/managed-infrastructure/ark/test/dev/busy',
    dockerfile: `
    FROM busybox
    WORKDIR /opt/app
    COPY ./ ./
    `,
  },
  sourceFiles: [
    './Tiltfile'
  ]
})

const deploy = ark.actions.deploy({
  name: 'deploy',
  attributes: {
    manifest: microservice({
      name: 'busy',
      serviceAccountName: 'busy',
      image: image.attributes.url,
      command: ['sh', '-c', '/usr/local/bin/ark-entrypoint tail -f /opt/app/src/dev/Tiltfile']
    }),
    liveSyncEnabled: true
  },
  dependsOn:[image]
})

import postgres from 'ark/plugins/@mantl/sre/postgres'
import postgresPlugin from '../go/lib/arkplugins/postgres/build'

const name = 'postgres'
ark.actions.deploy({
    name,
    attributes: {
        manifest: postgres({
            name,
        }),
    },
    dependsOn: [
        postgresPlugin,
    ]
})
