import * as ark from 'arksdk'
import vault from 'ark/plugins/@mantl/sre/vault'

export const deployVault = ark.actions.deploy({
  name: 'deploy',
  attributes: {
    manifest: vault({
      name: 'vault',
    }),
    portForward: {
      http: {
        remotePort: '8200',
        hostPort: '8200'
      }
    },
  },
})
