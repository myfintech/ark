module.exports = {
  title: 'Ark',
  tagline: 'OSS Documentation',
  url: 'https://[domain]',
  baseUrl: '/',
  favicon: 'img/favicon.ico',
  organizationName: 'myfintech',
  projectName: 'ark',
  themeConfig: {
    defaultDarkMode: true,
    navbar: {
      title: 'Ark OSS',
      logo: {
        alt: 'Github Logo',
        src: 'img/github.svg',
        // srcDark: 'img/github.svg',
      },
      links: [
        {
          to: 'docs/',
          activeBasePath: 'docs',
          label: 'Docs',
          position: 'left',
        },
        {to: 'blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/myfintech/ark',
          // label: 'GitHub',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
    footer: {
      // style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Style Guide',
              to: 'docs/',
            },
            {
              label: 'Second Doc',
              to: 'docs/doc2/',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Stack Overflow',
              href: 'https://stackoverflow.com/questions/tagged/arkbuild',
            },
            // {
            //   label: 'Discord',
            //   href: 'https://discordapp.com/invite/ark.build',
            // },
            // {
            //   label: 'Twitter',
            //   href: 'https://twitter.com/ark.build',
            // },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'Blog',
              to: 'blog',
            },
            {
              label: 'GitHub',
              href: 'https://github.com/facebook/docusaurus',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Fin Technologies, Inc. Built with Docusaurus.`,
    },
    prism: {
      additionalLanguages: ['docker', 'hcl'],
      plugins: ['line-numbers', 'show-language'],
      theme: require('@kiwicopple/prism-react-renderer/themes/vsDark'),
      darkTheme: require('@kiwicopple/prism-react-renderer/themes/vsDark'),
    },
    // TODO: enable searching
    // Requires that we run the DOCSEARCH indexer our selves since we're behind cloudflare
    // https://docsearch.algolia.com/docs/faq
    // https://docsearch.algolia.com/docs/config-file/
    // https://docsearch.algolia.com/docs/run-your-own
    // algolia: {
    //   apiKey: 'api-key',
    //   indexName: 'index-name',
    //   appId: 'app-id', // Optional, if you run the DocSearch crawler on your own
    //   algoliaOptions: {}, // Optional, if provided by Algolia
    // },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          // It is recommended to set document id as docs home page (`docs/` path).
          homePageId: 'ark/quick-start',
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl:
              'https://github.com/myfintech/ark/edit/main/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
              'https://github.com/myfintech/ark/edit/main/website/blog/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
