# Changelog

## [0.4.0](https://github.com/josephschmitt/monocle/compare/v0.3.0...v0.4.0) (2026-03-21)


### Features

* **tui:** add horizontal scrolling, line wrapping, and fix border width ([cc4356a](https://github.com/josephschmitt/monocle/commit/cc4356a3d68198179552c4c82f8551a8c855fb34))

## [0.3.0](https://github.com/josephschmitt/monocle/compare/v0.2.0...v0.3.0) (2026-03-20)


### Features

* **tui:** add responsive stacked layout for narrow terminals ([e9b6e3d](https://github.com/josephschmitt/monocle/commit/e9b6e3d52046c4dabcbbdfb6010f780ec4287dda))
* **tui:** add syntax highlighting and intra-line diff to diff view ([d291a30](https://github.com/josephschmitt/monocle/commit/d291a30c4505abcd1a238960b6be61ff304851fe))
* **tui:** add viewport scrolling to sidebar and cross-panel J/K diff scrolling ([8034f48](https://github.com/josephschmitt/monocle/commit/8034f484f82f9b9e55550afb2aeec36d26c5da63))


### Bug Fixes

* configure release-please to update README version strings ([b5f3a29](https://github.com/josephschmitt/monocle/commit/b5f3a298053e36cc10befb518930ef6fef3ce89c))

## [0.2.0](https://github.com/josephschmitt/monocle/compare/v0.1.0...v0.2.0) (2026-03-20)


### ⚠ BREAKING CHANGES

* CLI subcommands start, resume, and sessions have been removed. The --agent flag is gone. Just run `monocle` to start.
* CLI subcommands review-status, get-feedback, and submit-content have been removed. Use the MCP channel instead.
* All hook-related APIs removed. Skills replace hooks entirely.

### Features

* **adapters:** add --global flag for user-level .mcp.json install ([0af41b6](https://github.com/josephschmitt/monocle/commit/0af41b64d66246e2d77e4dd420739c0dacd02f52))
* **adapters:** add MCP channel server and installation for Claude Code ([0236e00](https://github.com/josephschmitt/monocle/commit/0236e00ec8107c87be9132a2533d96152250cd59))
* add install/uninstall commands with multi-agent hook management ([4327fa5](https://github.com/josephschmitt/monocle/commit/4327fa5dc889fe7a4309c5ca62fa27d66a1e96d8))
* auto-approve stop hook when nothing to review and inject plan content ([ba8571d](https://github.com/josephschmitt/monocle/commit/ba8571da2e98f9899cbb33d5a813e0a1ffbd5ae6))
* **core:** add persistent subscription support to socket server ([0c3b71f](https://github.com/josephschmitt/monocle/commit/0c3b71f6a65cbc11051e7c3801338514fff575f2))
* deterministic socket routing for multi-instance support ([82848cf](https://github.com/josephschmitt/monocle/commit/82848cfe83ae683f926920d11bdc9a05204b4693))
* make wait-for-review the primary skill flow ([249ece1](https://github.com/josephschmitt/monocle/commit/249ece145a7aac02b96b6461ca9ec13b3b2166a4))
* **protocol:** add subscribe and event notification message types ([d058a38](https://github.com/josephschmitt/monocle/commit/d058a38c44f8b52d9e70be8c3c5d6e6ff230d76a))
* replace hook-based agent integration with skills ([8ec3553](https://github.com/josephschmitt/monocle/commit/8ec355399389c5530b396813891d6f90f1d56486))
* strengthen skill prompt to check feedback more aggressively ([462afc5](https://github.com/josephschmitt/monocle/commit/462afc5551fb0f07ccc7cdb2b7f16e9499478ff3))
* **tui:** add collapsible tree view for files sidebar ([5c83132](https://github.com/josephschmitt/monocle/commit/5c831325daba3b516918fa1edccc44b5ee175e8d))
* **tui:** auto-advance base ref and add ref picker modal ([c59453b](https://github.com/josephschmitt/monocle/commit/c59453b856eeb55e5464d74ce2616e2fb4602580))


### Bug Fixes

* **adapters:** use correct MCP channel API and install deps ([1bde7a0](https://github.com/josephschmitt/monocle/commit/1bde7a06d7b86d90f2eca90d2ae4a2b54dcd3abc))
* advance baseRef on review round so file pane resets between rounds ([5757790](https://github.com/josephschmitt/monocle/commit/5757790ec7ce65161b93458678ca757a23b5a2b5))
* **tui:** auto-select content item when no files to review ([5940856](https://github.com/josephschmitt/monocle/commit/5940856c07f5ff3457971190791f6f45cefe55fd))
* **tui:** auto-select file when current view is stale or content ([0f831b0](https://github.com/josephschmitt/monocle/commit/0f831b08be8095f17ef8fcdb15ed2ee36c216b33))
* **tui:** auto-select first file when new files appear in empty view ([3c7e704](https://github.com/josephschmitt/monocle/commit/3c7e704516eb940c246dbde4910cabdd87ae6983))
* **tui:** auto-select from refreshResultMsg when view is stale ([24a6aa1](https://github.com/josephschmitt/monocle/commit/24a6aa142502e2641df9d1474bd5b05e44111a6f))
* **tui:** color ref picker hashes and prevent plan stealing focus ([4529489](https://github.com/josephschmitt/monocle/commit/4529489ec1b8af01f0565d7bf52344ff3d64947c))
* **tui:** fix space key in comment editor and use enter to save ([3af8f19](https://github.com/josephschmitt/monocle/commit/3af8f1961b27ce71cf55558a60120e6a25c62102))
* **tui:** left-align line numbers in content view gutter ([9485448](https://github.com/josephschmitt/monocle/commit/948544835f27fc0d48519ec2ecb1f3bf43817e2a))
* **tui:** prevent refresh tick from clobbering content view ([e5b51a6](https://github.com/josephschmitt/monocle/commit/e5b51a6ac8783b7f6a513d502257dd4f52886f55))
* **tui:** route loadContentMsg to diffView in app Update ([5b16a5b](https://github.com/josephschmitt/monocle/commit/5b16a5b9dd7a13f0df9cd6318e65a3197041ba9c))
* **tui:** use lowercase b for ref picker keybinding ([b5d02bd](https://github.com/josephschmitt/monocle/commit/b5d02bdac04a80b2c8d060306b9580b7e5cb2bd1))
* **tui:** use single-column line numbers for content view ([60a0cce](https://github.com/josephschmitt/monocle/commit/60a0cce5e3eca0496b923e7d6fd321841c25621c))


### Code Refactoring

* remove skill-based model, go channel-only ([24cb45f](https://github.com/josephschmitt/monocle/commit/24cb45fbc6a85e5925c08651d81bc245269c7ab7))
* update language, docs, and CLI for MCP channel model ([53d3b66](https://github.com/josephschmitt/monocle/commit/53d3b6607626015f56b1bba18de28e4ee53f8214))
