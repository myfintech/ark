import * as ark from 'arksdk'
import * as filepath from 'arksdk/filepath'
import config from '../../../build.config'
import * as pkginfo from '../../go/tools/pkginfo/build'
import * as checksum from '../../go/tools/checksum/build'

// language=docker
const dockerTemplate = `
FROM ${checksum.bin.attributes.url} as checksum
FROM ${pkginfo.bin.attributes.url} as pkginfo
FROM golang:1.16-alpine as build
COPY --from=checksum /checksum-linux /bin/checksum
COPY --from=pkginfo /pkginfo-linux /bin/pkginfo
`

export const bin = ark.actions.buildDockerImage({
  name: 'bin',
  dependsOn: [pkginfo.bin, checksum.bin],
  attributes: {
    repo: `${config.baseRepo}/go.1.16-alpine`,
    dockerfile: dockerTemplate
  }
})