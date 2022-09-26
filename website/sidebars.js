const fs = require('fs')
const path = require('path')
const docsDir = path.join(__dirname, 'docs')

module.exports = {
  someSidebar: {
    ARK: [
      'ark/quick-start',
      {
        Concepts: [
          'ark/concepts/overview',
          {
            Configuration: [
              'ark/concepts/configuration/global-attribute-reference',
              'ark/concepts/configuration/workspace-hcl-reference',
              'ark/concepts/configuration/build-hcl-reference',
              {
                Targets: dirToSidebar(docsDir, 'ark/concepts/configuration/targets/')
              }
            ]
          },
        ]
      },
      {
        Language: [
          'ark/language/expressions',
          {
            Functions: dirToSidebar(docsDir, 'ark/language/functions/')
          }
        ]
      },
      {
        'External Resources': [
          {
            type: 'link',
            label: 'Reproducible builds ',
            href: 'https://reproducible-builds.org/docs/buy-in/'
          },
          {
            type: 'link',
            label: 'Deterministic build systems',
            href: 'https://reproducible-builds.org/docs/deterministic-build-systems/'
          },
        ]
      }
    ],
  },
}

function dirToSidebar (basePath, prefix) {
  return fs.readdirSync(path.join(basePath, prefix)).map(p => path.join(prefix, path.basename(p, '.md')))
}
