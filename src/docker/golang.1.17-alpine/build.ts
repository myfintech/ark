import * as ark from 'arksdk'
import config from '../../../build.config'

// language=docker
const dockerTemplate = `
FROM golang:1.17-alpine as build
`

export const bin = ark.actions.buildDockerImage({
  name: 'bin',
  attributes: {
    repo: `${config.baseRepo}/golang.1.17-alpine`,
    dockerfile: dockerTemplate
  }
})