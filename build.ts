import * as ark from 'arksdk'
import config from './build.config'

export const goModules = ark.actions.buildDockerImage({
  name: 'go-modules',
  sourceFiles: ['./go.mod', './go.sum'],
  attributes: {
    repo: `${config.baseRepo}/go-modules`,
    dockerfile: `
    FROM golang:1.16-alpine as build

    ENV CGO_ENABLED=0
    
    WORKDIR /opt/app
    COPY ./ ./
    
    RUN apk add --no-cache git
    RUN go mod download
   `
  }
})