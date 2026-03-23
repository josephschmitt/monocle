# Changelog

## [0.16.0](https://github.com/josephschmitt/monocle/compare/v0.15.0...v0.16.0) (2026-03-23)


### Features

* **tui:** add clear_after_submit config option with session override ([d18620c](https://github.com/josephschmitt/monocle/commit/d18620c9faaf173cb1e71923a7eb40581b9410f4))

## [0.15.0](https://github.com/josephschmitt/monocle/compare/v0.14.1...v0.15.0) (2026-03-23)


### ⚠ BREAKING CHANGES

* `monocle install` is replaced by `monocle register`. Existing .mcp.json entries will be detected as outdated and users will be prompted to re-register.
* CLI subcommands start, resume, and sessions have been removed. The --agent flag is gone. Just run `monocle` to start.
* CLI subcommands review-status, get-feedback, and submit-content have been removed. Use the MCP channel instead.
* All hook-related APIs removed. Skills replace hooks entirely.

### Features

* **adapters:** add --global flag for user-level .mcp.json install ([0af41b6](https://github.com/josephschmitt/monocle/commit/0af41b64d66246e2d77e4dd420739c0dacd02f52))
* **adapters:** add MCP channel server and installation for Claude Code ([0236e00](https://github.com/josephschmitt/monocle/commit/0236e00ec8107c87be9132a2533d96152250cd59))
* **adapters:** auto-detect JavaScript runtime for MCP channel ([b34028f](https://github.com/josephschmitt/monocle/commit/b34028f2ca35cd620ff9a36769733b8d4043b61a))
* add --version flag with goreleaser-injected version ([3fb32d2](https://github.com/josephschmitt/monocle/commit/3fb32d29b914677be16f6b9ab15763e683f02413))
* add config settings for layout, diff style, wrap, tab size, and context lines ([f10848e](https://github.com/josephschmitt/monocle/commit/f10848e616cccb75d67819e94305da5f433f9a7e))
* add content_type param to submit_plan for syntax highlighting ([315c934](https://github.com/josephschmitt/monocle/commit/315c9341eb60ec70bec8a8b40e75c44e4d961964))
* add install/uninstall commands with multi-agent hook management ([4327fa5](https://github.com/josephschmitt/monocle/commit/4327fa5dc889fe7a4309c5ca62fa27d66a1e96d8))
* auto-approve stop hook when nothing to review and inject plan content ([ba8571d](https://github.com/josephschmitt/monocle/commit/ba8571da2e98f9899cbb33d5a813e0a1ffbd5ae6))
* auto-detect and offer MCP channel install on TUI launch ([8685e52](https://github.com/josephschmitt/monocle/commit/8685e52b2dea06c2f0ea3661f55a87a5e9c06335))
* **core:** add persistent subscription support to socket server ([0c3b71f](https://github.com/josephschmitt/monocle/commit/0c3b71f6a65cbc11051e7c3801338514fff575f2))
* **core:** wire ReviewFormatConfig into formatter ([4272f79](https://github.com/josephschmitt/monocle/commit/4272f7932db745731a2f7f808093473f37b35181))
* deterministic socket routing for multi-instance support ([82848cf](https://github.com/josephschmitt/monocle/commit/82848cfe83ae683f926920d11bdc9a05204b4693))
* implement comment resolution flow ([dffb57b](https://github.com/josephschmitt/monocle/commit/dffb57bf97602313421e236f30d86c8611091913))
* make wait-for-review the primary skill flow ([249ece1](https://github.com/josephschmitt/monocle/commit/249ece145a7aac02b96b6461ca9ec13b3b2166a4))
* **protocol:** add subscribe and event notification message types ([d058a38](https://github.com/josephschmitt/monocle/commit/d058a38c44f8b52d9e70be8c3c5d6e6ff230d76a))
* replace hook-based agent integration with skills ([8ec3553](https://github.com/josephschmitt/monocle/commit/8ec355399389c5530b396813891d6f90f1d56486))
* replace install/uninstall with register/unregister and serve-mcp-channel ([2d21e6e](https://github.com/josephschmitt/monocle/commit/2d21e6e95ae94c67bbe1c119ef6dc128ef7d1799))
* strengthen skill prompt to check feedback more aggressively ([462afc5](https://github.com/josephschmitt/monocle/commit/462afc5551fb0f07ccc7cdb2b7f16e9499478ff3))
* **tui:** add collapsible tree view for files sidebar ([5c83132](https://github.com/josephschmitt/monocle/commit/5c831325daba3b516918fa1edccc44b5ee175e8d))
* **tui:** add connection indicator, info modal, and manual socket override ([6696954](https://github.com/josephschmitt/monocle/commit/6696954de25f237e2dfd4590b37c1f3fc5e8aec8))
* **tui:** add copy review to clipboard in submit modal ([8e37677](https://github.com/josephschmitt/monocle/commit/8e376772a6a6515d3d327bec1a18481f06eb7432))
* **tui:** add file-level commenting with C key ([8f00d5e](https://github.com/josephschmitt/monocle/commit/8f00d5e234ed6be5d9082bb29379f2dff26b8766))
* **tui:** add horizontal scrolling, line wrapping, and fix border width ([cc4356a](https://github.com/josephschmitt/monocle/commit/cc4356a3d68198179552c4c82f8551a8c855fb34))
* **tui:** add line-preserving markdown styling for plans ([47b8774](https://github.com/josephschmitt/monocle/commit/47b8774342fd267adfc2c6396e2e915c1085307e))
* **tui:** add o_(◉) ASCII logo to title bar ([5f5855e](https://github.com/josephschmitt/monocle/commit/5f5855e39b4f57075b392a8a25789648b3520919))
* **tui:** add responsive stacked layout for narrow terminals ([e9b6e3d](https://github.com/josephschmitt/monocle/commit/e9b6e3d52046c4dabcbbdfb6010f780ec4287dda))
* **tui:** add splash screen with setup instructions and keybinding hints ([398902c](https://github.com/josephschmitt/monocle/commit/398902cec26ef3db5d876bdc4fbc951808b69ad9))
* **tui:** add submission history view ([b624109](https://github.com/josephschmitt/monocle/commit/b624109594eee000cdcbf3f899b9072a19e853a0))
* **tui:** add syntax highlighting and intra-line diff to diff view ([d291a30](https://github.com/josephschmitt/monocle/commit/d291a30c4505abcd1a238960b6be61ff304851fe))
* **tui:** add viewport scrolling to sidebar and cross-panel J/K diff scrolling ([8034f48](https://github.com/josephschmitt/monocle/commit/8034f484f82f9b9e55550afb2aeec36d26c5da63))
* **tui:** apply markdown styling to changed markdown files in diff view ([70b0408](https://github.com/josephschmitt/monocle/commit/70b04087468868c7e79fbeac2ad13e0d20a99fa0))
* **tui:** auto-advance base ref and add ref picker modal ([c59453b](https://github.com/josephschmitt/monocle/commit/c59453b856eeb55e5464d74ce2616e2fb4602580))
* **tui:** clear comments on submit, discard command, and review status selector ([4b3e058](https://github.com/josephschmitt/monocle/commit/4b3e058b472061b410e46a1d1eefb2dd589e5a4c))
* **tui:** contextual comment keybinds and status bar hints ([57079f3](https://github.com/josephschmitt/monocle/commit/57079f30a8bdd763d685c7f3df996cdebfbaac62))
* **tui:** cross-pane file navigation, half-page scroll, and unfocused selection indicator ([f307eaa](https://github.com/josephschmitt/monocle/commit/f307eaad16be860fe19875b14cc6da0f957f047d))
* **tui:** implement configurable keybindings system ([a4f0558](https://github.com/josephschmitt/monocle/commit/a4f0558044c843e35742cf3fd8449031e35740f9))
* **tui:** improve ref picker with scrolling, pre-selection, and load more ([522daa0](https://github.com/josephschmitt/monocle/commit/522daa0d8e15629098ef116a4865daa2011f0e25))
* **tui:** persist sidebar style preference across sessions ([42489f9](https://github.com/josephschmitt/monocle/commit/42489f961e82c65a8e7ab893f4f0b53a455e9023))
* **tui:** raise layout breakpoint and prioritize diff area width ([2aa1ba4](https://github.com/josephschmitt/monocle/commit/2aa1ba4494a2f88d5db8895d017fe539b34f3bf1))
* **tui:** replace confirm modal with dedicated install prompt supporting global/local scope ([d5c2e94](https://github.com/josephschmitt/monocle/commit/d5c2e942235e6d983c00856801de391d593ffaad))
* **tui:** style comment type selector with colored pill tabs ([43776c8](https://github.com/josephschmitt/monocle/commit/43776c84a92618a538185750c384a0eb67852d79))
* **tui:** wrap at word boundaries instead of character boundaries ([8a550f6](https://github.com/josephschmitt/monocle/commit/8a550f6205f0b1a0025c07d9787033c029058cf9))


### Bug Fixes

* **adapters:** use correct MCP channel API and install deps ([1bde7a0](https://github.com/josephschmitt/monocle/commit/1bde7a06d7b86d90f2eca90d2ae4a2b54dcd3abc))
* advance baseRef on review round so file pane resets between rounds ([5757790](https://github.com/josephschmitt/monocle/commit/5757790ec7ce65161b93458678ca757a23b5a2b5))
* configure release-please to update README version strings ([b5f3a29](https://github.com/josephschmitt/monocle/commit/b5f3a298053e36cc10befb518930ef6fef3ce89c))
* **core:** fix off-by-one in base ref selection ([cad4deb](https://github.com/josephschmitt/monocle/commit/cad4deb7356877e0dc97b5fc5d3f2615aaebd9eb))
* **docs:** use ASCII arrows in flow diagram for consistent rendering ([49ed5c9](https://github.com/josephschmitt/monocle/commit/49ed5c943145dfedbf26e37b610b0239777e2371))
* ignore node_modules symlink in worktrees ([862b2bb](https://github.com/josephschmitt/monocle/commit/862b2bbeeaa54165125b0dd1c0ff2e391a768a7d))
* register event handlers before sending subscribe ack to prevent race ([64d4f81](https://github.com/josephschmitt/monocle/commit/64d4f81ffa209e8e1be45224371e0ebd8d973646))
* remove version from goreleaser archive names to fix release-please URLs ([041b270](https://github.com/josephschmitt/monocle/commit/041b270d170bc6dca15755511bdf51e8f9932899))
* restore incomplete features incorrectly removed as dead code ([e459357](https://github.com/josephschmitt/monocle/commit/e459357bbcd0a80126e357aa8c62fb9e2339080d))
* **test:** isolate setupTestRepo from parent worktree git environment ([2594164](https://github.com/josephschmitt/monocle/commit/2594164fbd5d7026cbbb900be208fbbf7c938e31))
* **test:** use git init -b to avoid branch name collision in worktrees ([ccbd564](https://github.com/josephschmitt/monocle/commit/ccbd564364e127aa7ac25ebfd97b22ee14db2232))
* **tui:** allow stacked sidebar to grow with terminal height ([78a507a](https://github.com/josephschmitt/monocle/commit/78a507a9529328beb34d7dfb2ca6e39cefe79972))
* **tui:** allow toggle review keybind to work in diff viewer ([12741df](https://github.com/josephschmitt/monocle/commit/12741df5e42d292f02c35b842a20c39f4e70405d))
* **tui:** auto-select content item when no files to review ([5940856](https://github.com/josephschmitt/monocle/commit/5940856c07f5ff3457971190791f6f45cefe55fd))
* **tui:** auto-select file when current view is stale or content ([0f831b0](https://github.com/josephschmitt/monocle/commit/0f831b08be8095f17ef8fcdb15ed2ee36c216b33))
* **tui:** auto-select first file when new files appear in empty view ([3c7e704](https://github.com/josephschmitt/monocle/commit/3c7e704516eb940c246dbde4910cabdd87ae6983))
* **tui:** auto-select from refreshResultMsg when view is stale ([24a6aa1](https://github.com/josephschmitt/monocle/commit/24a6aa142502e2641df9d1474bd5b05e44111a6f))
* **tui:** clear visual mode after saving a comment ([a4f4363](https://github.com/josephschmitt/monocle/commit/a4f4363175aeec743458cbbe4aab80112c5b8b58))
* **tui:** clear visual mode only on comment submit, not on reload ([26829ca](https://github.com/josephschmitt/monocle/commit/26829ca7ff9bb2a18bf5901064e42994063b8043))
* **tui:** color ref picker hashes and prevent plan stealing focus ([4529489](https://github.com/josephschmitt/monocle/commit/4529489ec1b8af01f0565d7bf52344ff3d64947c))
* **tui:** default review status based on comment types ([7ffc803](https://github.com/josephschmitt/monocle/commit/7ffc803d742c6ebd2bd648c74fc5d2c74a6b059a))
* **tui:** fix modal overlay breaking borders and improve modal sizing ([1b7b11b](https://github.com/josephschmitt/monocle/commit/1b7b11b6c295b4e8dc68d9d5d453b470813dc7a7))
* **tui:** fix plan review feedback flow and content view stability ([dda4bab](https://github.com/josephschmitt/monocle/commit/dda4bab279bcdfde660121626b0efb763de88dd9))
* **tui:** fix space key in comment editor and use enter to save ([3af8f19](https://github.com/josephschmitt/monocle/commit/3af8f1961b27ce71cf55558a60120e6a25c62102))
* **tui:** fix split diff layout overflow caused by tab characters ([a0d0382](https://github.com/josephschmitt/monocle/commit/a0d0382c091b9fffaea662da6afcf81a817c2e91))
* **tui:** guard against nil session in refreshFiles ([c2ee7ea](https://github.com/josephschmitt/monocle/commit/c2ee7ead18b2a0f98f32b68f12081ff5842e466f))
* **tui:** integrate markdown styling, inline comments, and scroll fixes ([86446d6](https://github.com/josephschmitt/monocle/commit/86446d6ef47fc5699f8bf0950e870494ae49ee8f))
* **tui:** left-align line numbers in content view gutter ([9485448](https://github.com/josephschmitt/monocle/commit/948544835f27fc0d48519ec2ecb1f3bf43817e2a))
* **tui:** make comment lines selectable so resolve keybind works ([adfc284](https://github.com/josephschmitt/monocle/commit/adfc2842f44a6321ae957bf8808598b6ecf83e58))
* **tui:** prevent refresh tick from clobbering content view ([e5b51a6](https://github.com/josephschmitt/monocle/commit/e5b51a6ac8783b7f6a513d502257dd4f52886f55))
* **tui:** recalculate stacked layout when files or content items change ([89c5b26](https://github.com/josephschmitt/monocle/commit/89c5b26532b5e958fbec4a7692bf7b8437fbcc29))
* **tui:** reduce modal top padding and add help modal scrolling ([eff7cb1](https://github.com/josephschmitt/monocle/commit/eff7cb1cbd30aac0c97e39ce0caf163659495b75))
* **tui:** render inline comments at target line with per-type colors ([87a6bde](https://github.com/josephschmitt/monocle/commit/87a6bde60fa5d3437a6a1010aa54416cf4835116))
* **tui:** route loadContentMsg to diffView in app Update ([5b16a5b](https://github.com/josephschmitt/monocle/commit/5b16a5b9dd7a13f0df9cd6318e65a3197041ba9c))
* **tui:** show devicons for content items and reorder sidebar sections ([dbc179c](https://github.com/josephschmitt/monocle/commit/dbc179c32a57d0097c0f75d589e4657f4665a5cd))
* **tui:** skip removed lines in cursor selection ([66ba07a](https://github.com/josephschmitt/monocle/commit/66ba07a22cf325f8a9984e8278e9cad5cc8eb0c9))
* **tui:** use lowercase b for ref picker keybinding ([b5d02bd](https://github.com/josephschmitt/monocle/commit/b5d02bdac04a80b2c8d060306b9580b7e5cb2bd1))
* **tui:** use OSC 52 for clipboard and fix yank keybind casing ([a12e99b](https://github.com/josephschmitt/monocle/commit/a12e99b50e051020e0197211c14555dfe1148f03))
* **tui:** use single-column line numbers for content view ([60a0cce](https://github.com/josephschmitt/monocle/commit/60a0cce5e3eca0496b923e7d6fd321841c25621c))
* use ${HOME} in .mcp.json channel path instead of absolute path ([372ba71](https://github.com/josephschmitt/monocle/commit/372ba71f25be1dc0f22e4cefe91fdda2d591c5af))
* wire ContentItemProvider on formatter for plan content snippets ([daa02eb](https://github.com/josephschmitt/monocle/commit/daa02eb0095c6d3b72ede0514527f3ca77c465bf))
* wrap goreleaser before hook in sh -c for shell builtin support ([48c1746](https://github.com/josephschmitt/monocle/commit/48c1746b03884c6c7518aed718eddb30ca0c5d5f))


### Code Refactoring

* remove skill-based model, go channel-only ([24cb45f](https://github.com/josephschmitt/monocle/commit/24cb45fbc6a85e5925c08651d81bc245269c7ab7))
* update language, docs, and CLI for MCP channel model ([53d3b66](https://github.com/josephschmitt/monocle/commit/53d3b6607626015f56b1bba18de28e4ee53f8214))

## [0.14.1](https://github.com/josephschmitt/monocle/compare/v0.14.0...v0.14.1) (2026-03-23)


### Bug Fixes

* wrap goreleaser before hook in sh -c for shell builtin support ([48c1746](https://github.com/josephschmitt/monocle/commit/48c1746b03884c6c7518aed718eddb30ca0c5d5f))

## [0.14.0](https://github.com/josephschmitt/monocle/compare/v0.13.0...v0.14.0) (2026-03-23)


### ⚠ BREAKING CHANGES

* `monocle install` is replaced by `monocle register`. Existing .mcp.json entries will be detected as outdated and users will be prompted to re-register.

### Features

* replace install/uninstall with register/unregister and serve-mcp-channel ([2d21e6e](https://github.com/josephschmitt/monocle/commit/2d21e6e95ae94c67bbe1c119ef6dc128ef7d1799))

## [0.13.0](https://github.com/josephschmitt/monocle/compare/v0.12.0...v0.13.0) (2026-03-22)


### Features

* **tui:** improve ref picker with scrolling, pre-selection, and load more ([522daa0](https://github.com/josephschmitt/monocle/commit/522daa0d8e15629098ef116a4865daa2011f0e25))

## [0.12.0](https://github.com/josephschmitt/monocle/compare/v0.11.0...v0.12.0) (2026-03-22)


### Features

* **core:** wire ReviewFormatConfig into formatter ([4272f79](https://github.com/josephschmitt/monocle/commit/4272f7932db745731a2f7f808093473f37b35181))
* implement comment resolution flow ([dffb57b](https://github.com/josephschmitt/monocle/commit/dffb57bf97602313421e236f30d86c8611091913))
* **tui:** add submission history view ([b624109](https://github.com/josephschmitt/monocle/commit/b624109594eee000cdcbf3f899b9072a19e853a0))
* **tui:** contextual comment keybinds and status bar hints ([57079f3](https://github.com/josephschmitt/monocle/commit/57079f30a8bdd763d685c7f3df996cdebfbaac62))
* **tui:** implement configurable keybindings system ([a4f0558](https://github.com/josephschmitt/monocle/commit/a4f0558044c843e35742cf3fd8449031e35740f9))


### Bug Fixes

* restore incomplete features incorrectly removed as dead code ([e459357](https://github.com/josephschmitt/monocle/commit/e459357bbcd0a80126e357aa8c62fb9e2339080d))
* **test:** isolate setupTestRepo from parent worktree git environment ([2594164](https://github.com/josephschmitt/monocle/commit/2594164fbd5d7026cbbb900be208fbbf7c938e31))
* **test:** use git init -b to avoid branch name collision in worktrees ([ccbd564](https://github.com/josephschmitt/monocle/commit/ccbd564364e127aa7ac25ebfd97b22ee14db2232))
* **tui:** guard against nil session in refreshFiles ([c2ee7ea](https://github.com/josephschmitt/monocle/commit/c2ee7ead18b2a0f98f32b68f12081ff5842e466f))
* **tui:** make comment lines selectable so resolve keybind works ([adfc284](https://github.com/josephschmitt/monocle/commit/adfc2842f44a6321ae957bf8808598b6ecf83e58))

## [0.11.0](https://github.com/josephschmitt/monocle/compare/v0.10.1...v0.11.0) (2026-03-22)


### Features

* add --version flag with goreleaser-injected version ([3fb32d2](https://github.com/josephschmitt/monocle/commit/3fb32d29b914677be16f6b9ab15763e683f02413))
* **tui:** add connection indicator, info modal, and manual socket override ([6696954](https://github.com/josephschmitt/monocle/commit/6696954de25f237e2dfd4590b37c1f3fc5e8aec8))
* **tui:** add copy review to clipboard in submit modal ([8e37677](https://github.com/josephschmitt/monocle/commit/8e376772a6a6515d3d327bec1a18481f06eb7432))
* **tui:** add line-preserving markdown styling for plans ([47b8774](https://github.com/josephschmitt/monocle/commit/47b8774342fd267adfc2c6396e2e915c1085307e))
* **tui:** apply markdown styling to changed markdown files in diff view ([70b0408](https://github.com/josephschmitt/monocle/commit/70b04087468868c7e79fbeac2ad13e0d20a99fa0))
* **tui:** wrap at word boundaries instead of character boundaries ([8a550f6](https://github.com/josephschmitt/monocle/commit/8a550f6205f0b1a0025c07d9787033c029058cf9))


### Bug Fixes

* remove version from goreleaser archive names to fix release-please URLs ([041b270](https://github.com/josephschmitt/monocle/commit/041b270d170bc6dca15755511bdf51e8f9932899))
* **tui:** clear visual mode after saving a comment ([a4f4363](https://github.com/josephschmitt/monocle/commit/a4f4363175aeec743458cbbe4aab80112c5b8b58))
* **tui:** clear visual mode only on comment submit, not on reload ([26829ca](https://github.com/josephschmitt/monocle/commit/26829ca7ff9bb2a18bf5901064e42994063b8043))
* **tui:** fix plan review feedback flow and content view stability ([dda4bab](https://github.com/josephschmitt/monocle/commit/dda4bab279bcdfde660121626b0efb763de88dd9))
* **tui:** integrate markdown styling, inline comments, and scroll fixes ([86446d6](https://github.com/josephschmitt/monocle/commit/86446d6ef47fc5699f8bf0950e870494ae49ee8f))
* **tui:** recalculate stacked layout when files or content items change ([89c5b26](https://github.com/josephschmitt/monocle/commit/89c5b26532b5e958fbec4a7692bf7b8437fbcc29))
* **tui:** show devicons for content items and reorder sidebar sections ([dbc179c](https://github.com/josephschmitt/monocle/commit/dbc179c32a57d0097c0f75d589e4657f4665a5cd))
* **tui:** use OSC 52 for clipboard and fix yank keybind casing ([a12e99b](https://github.com/josephschmitt/monocle/commit/a12e99b50e051020e0197211c14555dfe1148f03))
* wire ContentItemProvider on formatter for plan content snippets ([daa02eb](https://github.com/josephschmitt/monocle/commit/daa02eb0095c6d3b72ede0514527f3ca77c465bf))

## [0.10.1](https://github.com/josephschmitt/monocle/compare/v0.10.0...v0.10.1) (2026-03-22)


### Bug Fixes

* register event handlers before sending subscribe ack to prevent race ([64d4f81](https://github.com/josephschmitt/monocle/commit/64d4f81ffa209e8e1be45224371e0ebd8d973646))
* **tui:** allow stacked sidebar to grow with terminal height ([78a507a](https://github.com/josephschmitt/monocle/commit/78a507a9529328beb34d7dfb2ca6e39cefe79972))
* **tui:** allow toggle review keybind to work in diff viewer ([12741df](https://github.com/josephschmitt/monocle/commit/12741df5e42d292f02c35b842a20c39f4e70405d))

## [0.10.0](https://github.com/josephschmitt/monocle/compare/v0.9.0...v0.10.0) (2026-03-22)


### Features

* auto-detect and offer MCP channel install on TUI launch ([8685e52](https://github.com/josephschmitt/monocle/commit/8685e52b2dea06c2f0ea3661f55a87a5e9c06335))
* **tui:** replace confirm modal with dedicated install prompt supporting global/local scope ([d5c2e94](https://github.com/josephschmitt/monocle/commit/d5c2e942235e6d983c00856801de391d593ffaad))


### Bug Fixes

* use ${HOME} in .mcp.json channel path instead of absolute path ([372ba71](https://github.com/josephschmitt/monocle/commit/372ba71f25be1dc0f22e4cefe91fdda2d591c5af))

## [0.9.0](https://github.com/josephschmitt/monocle/compare/v0.8.0...v0.9.0) (2026-03-21)


### Features

* add content_type param to submit_plan for syntax highlighting ([315c934](https://github.com/josephschmitt/monocle/commit/315c9341eb60ec70bec8a8b40e75c44e4d961964))

## [0.8.0](https://github.com/josephschmitt/monocle/compare/v0.7.0...v0.8.0) (2026-03-21)


### Features

* add config settings for layout, diff style, wrap, tab size, and context lines ([f10848e](https://github.com/josephschmitt/monocle/commit/f10848e616cccb75d67819e94305da5f433f9a7e))

## [0.7.0](https://github.com/josephschmitt/monocle/compare/v0.6.0...v0.7.0) (2026-03-21)


### Features

* **tui:** add splash screen with setup instructions and keybinding hints ([398902c](https://github.com/josephschmitt/monocle/commit/398902cec26ef3db5d876bdc4fbc951808b69ad9))
* **tui:** clear comments on submit, discard command, and review status selector ([4b3e058](https://github.com/josephschmitt/monocle/commit/4b3e058b472061b410e46a1d1eefb2dd589e5a4c))
* **tui:** cross-pane file navigation, half-page scroll, and unfocused selection indicator ([f307eaa](https://github.com/josephschmitt/monocle/commit/f307eaad16be860fe19875b14cc6da0f957f047d))
* **tui:** persist sidebar style preference across sessions ([42489f9](https://github.com/josephschmitt/monocle/commit/42489f961e82c65a8e7ab893f4f0b53a455e9023))
* **tui:** raise layout breakpoint and prioritize diff area width ([2aa1ba4](https://github.com/josephschmitt/monocle/commit/2aa1ba4494a2f88d5db8895d017fe539b34f3bf1))


### Bug Fixes

* ignore node_modules symlink in worktrees ([862b2bb](https://github.com/josephschmitt/monocle/commit/862b2bbeeaa54165125b0dd1c0ff2e391a768a7d))
* **tui:** default review status based on comment types ([7ffc803](https://github.com/josephschmitt/monocle/commit/7ffc803d742c6ebd2bd648c74fc5d2c74a6b059a))
* **tui:** reduce modal top padding and add help modal scrolling ([eff7cb1](https://github.com/josephschmitt/monocle/commit/eff7cb1cbd30aac0c97e39ce0caf163659495b75))

## [0.6.0](https://github.com/josephschmitt/monocle/compare/v0.5.0...v0.6.0) (2026-03-21)


### Features

* **adapters:** auto-detect JavaScript runtime for MCP channel ([b34028f](https://github.com/josephschmitt/monocle/commit/b34028f2ca35cd620ff9a36769733b8d4043b61a))


### Bug Fixes

* **docs:** use ASCII arrows in flow diagram for consistent rendering ([49ed5c9](https://github.com/josephschmitt/monocle/commit/49ed5c943145dfedbf26e37b610b0239777e2371))

## [0.5.0](https://github.com/josephschmitt/monocle/compare/v0.4.0...v0.5.0) (2026-03-21)


### Features

* **tui:** add file-level commenting with C key ([8f00d5e](https://github.com/josephschmitt/monocle/commit/8f00d5e234ed6be5d9082bb29379f2dff26b8766))
* **tui:** add o_(◉) ASCII logo to title bar ([5f5855e](https://github.com/josephschmitt/monocle/commit/5f5855e39b4f57075b392a8a25789648b3520919))
* **tui:** style comment type selector with colored pill tabs ([43776c8](https://github.com/josephschmitt/monocle/commit/43776c84a92618a538185750c384a0eb67852d79))


### Bug Fixes

* **core:** fix off-by-one in base ref selection ([cad4deb](https://github.com/josephschmitt/monocle/commit/cad4deb7356877e0dc97b5fc5d3f2615aaebd9eb))
* **tui:** fix modal overlay breaking borders and improve modal sizing ([1b7b11b](https://github.com/josephschmitt/monocle/commit/1b7b11b6c295b4e8dc68d9d5d453b470813dc7a7))
* **tui:** fix split diff layout overflow caused by tab characters ([a0d0382](https://github.com/josephschmitt/monocle/commit/a0d0382c091b9fffaea662da6afcf81a817c2e91))
* **tui:** render inline comments at target line with per-type colors ([87a6bde](https://github.com/josephschmitt/monocle/commit/87a6bde60fa5d3437a6a1010aa54416cf4835116))
* **tui:** skip removed lines in cursor selection ([66ba07a](https://github.com/josephschmitt/monocle/commit/66ba07a22cf325f8a9984e8278e9cad5cc8eb0c9))

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
