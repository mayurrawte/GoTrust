# GoTrust Wiki

This directory contains the documentation for GoTrust that will be displayed in the GitHub Wiki.

## Setting Up the Wiki

To publish these docs to your GitHub Wiki:

### Option 1: Manual Upload

1. Go to your repository: https://github.com/mayurrawte/GoTrust
2. Click on "Wiki" tab
3. Click "Create the first page" if wiki is empty
4. Copy content from each .md file here to create wiki pages

### Option 2: Using Git (Recommended)

1. Clone the wiki repository:
```bash
git clone https://github.com/mayurrawte/GoTrust.wiki.git
```

2. Copy all markdown files from this directory:
```bash
cp /Users/mayurrawte/thepsygeek/gotrust/wiki/*.md GoTrust.wiki/
```

3. Commit and push:
```bash
cd GoTrust.wiki
git add .
git commit -m "Add comprehensive documentation"
git push
```

## Wiki Structure

The wiki is organized into these sections:

### Core Documentation
- **Home.md** - Main landing page
- **Getting-Started.md** - Quick start guide
- **Framework-Adapters.md** - Framework integration guide
- **Database-Integration.md** - Database setup examples
- **FAQ.md** - Frequently asked questions

### Navigation
- **_Sidebar.md** - Sidebar navigation (appears on every page)
- **_Footer.md** - Footer content (optional)

## Maintaining the Wiki

### Adding New Pages

1. Create a new .md file in this directory
2. Add link to _Sidebar.md
3. Push to wiki repository

### Updating Existing Pages

1. Edit the .md file
2. Push changes to wiki repository

### Best Practices

- Use descriptive page names (use hyphens for spaces)
- Keep pages focused on single topics
- Include code examples
- Add links between related pages
- Update the sidebar when adding new pages

## Automation

You can automate wiki updates using GitHub Actions:

```yaml
name: Sync Wiki
on:
  push:
    paths:
      - 'wiki/**'
    branches:
      - main

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Sync Wiki
        uses: SwiftDocOrg/github-wiki-publish-action@main
        with:
          path: wiki
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Contributing to Documentation

We welcome documentation improvements! To contribute:

1. Fork the repository
2. Edit/add wiki files in the `wiki/` directory
3. Submit a pull request
4. We'll review and merge your changes

## License

The documentation is licensed under the same MIT license as GoTrust.