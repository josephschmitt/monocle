# Changelog

## [0.38.0](https://github.com/josephschmitt/monocle/compare/v0.37.0...v0.38.0) (2026-03-30)


### ⚠ BREAKING CHANGES

* **register:** embed SKILL.md files, drop MCP for non-channel agents
* **channel:** strip MCP tools from channel, move to CLI skills

### Features

* **cli:** add `monocle status` command and gate skills on running state ([169d053](https://github.com/josephschmitt/monocle/commit/169d0530e1ee96f9031ba0ca3510d1eb4b770c29))
* **cli:** add monocle review subcommands for agent-facing CLI tools ([d8fb933](https://github.com/josephschmitt/monocle/commit/d8fb9330abe7ad503767152e72e55139fc5161a8))
* **plugin:** make new claude plugin the default ([cb1f013](https://github.com/josephschmitt/monocle/commit/cb1f013c2a920db47815bd010d46a92d0a8dd5f6))
* **register:** auto-allow monocle permissions during agent registration ([5b5a1ff](https://github.com/josephschmitt/monocle/commit/5b5a1ff925faa072cf4afee2a1d76a6250cda232))
* **skills:** download skills from GitHub releases instead of embedding ([51e8758](https://github.com/josephschmitt/monocle/commit/51e87580186fcdb1d45901bdb8ba89f13be3bcd6))
* **skills:** sync root skills into plugin directories via make target ([a865a94](https://github.com/josephschmitt/monocle/commit/a865a9468f839468af72b11a22edf07d074ac971))
* **skills:** sync root skills into plugin directories via make target ([50fc456](https://github.com/josephschmitt/monocle/commit/50fc4569a7c053b035c5a66d488236b84c409179))
* **tui:** update splash screen to prefer `monocle register` for setup ([fae8aab](https://github.com/josephschmitt/monocle/commit/fae8aab01d4773b764d6c328da394d25b4359746))


### Bug Fixes

* **channel:** remove ListToolsRequestSchema handler that crashes without tools capability ([66bb0ae](https://github.com/josephschmitt/monocle/commit/66bb0aec42aae1775cde6462ee7ea903497cd2d1))
* **channel:** restore 10s engine connection wait after MCP handshake ([ed8d09c](https://github.com/josephschmitt/monocle/commit/ed8d09c994a8b0366db888f7a85896db32a8edf7))
* **ci:** move skills tarball out of dist/ to avoid goreleaser conflict ([b2b0dcd](https://github.com/josephschmitt/monocle/commit/b2b0dcd359981f3deb82535e86c4450a0045f0f5))


### Code Refactoring

* **channel:** strip MCP tools from channel, move to CLI skills ([b7e6402](https://github.com/josephschmitt/monocle/commit/b7e64022c42d17ae2f6bb26ca480ea221937982f))
* **register:** embed SKILL.md files, drop MCP for non-channel agents ([72ccb9e](https://github.com/josephschmitt/monocle/commit/72ccb9ee369cb06e573200056ceae2d69b156048))

## [0.37.0](https://github.com/josephschmitt/monocle/compare/v0.36.1...v0.37.0) (2026-03-30)


### Features

* **tui:** show yellow "Waiting for Review" in status bar when agent is blocked ([8eb31bd](https://github.com/josephschmitt/monocle/commit/8eb31bd37a9f0be098dff6af3aeafe05653512f2))


### Bug Fixes

* **tui:** make diff toggle a no-op on first-version plans ([ce5a00a](https://github.com/josephschmitt/monocle/commit/ce5a00a7b1ebdbab18226e49418544147d06b82f))
* **tui:** show correct delivery status in submit modal for queue-mode connections ([3afcd95](https://github.com/josephschmitt/monocle/commit/3afcd95070eca313b7aa9972dc81ac24ae2b8343))

## [0.36.1](https://github.com/josephschmitt/monocle/compare/v0.36.0...v0.36.1) (2026-03-30)


### Bug Fixes

* **tui:** preserve visual selection during file view refresh ([972192b](https://github.com/josephschmitt/monocle/commit/972192bae9ff36665566cd0fa47ecc82a2daff65))

## [0.36.0](https://github.com/josephschmitt/monocle/compare/v0.35.0...v0.36.0) (2026-03-30)


### Features

* **tui:** auto-hide sidebar when empty, auto-show when items arrive ([28d3edf](https://github.com/josephschmitt/monocle/commit/28d3edf342fead450514fa64227e76ebe96eb3dc))

## [0.35.0](https://github.com/josephschmitt/monocle/compare/v0.34.0...v0.35.0) (2026-03-30)


### Features

* add Codex and Gemini plugins, reorganize plugin structure ([1c482e8](https://github.com/josephschmitt/monocle/commit/1c482e8c3c5723bd6dd99ebf1373581083aa49f6))
* **plugin:** add side-by-side old and new plugin directories for testing ([d44db21](https://github.com/josephschmitt/monocle/commit/d44db2125aeb7604f825e6efe3a9cfd73d2611f2))


### Bug Fixes

* **tui:** prevent refresh timer from clobbering content diff view ([b39c555](https://github.com/josephschmitt/monocle/commit/b39c555fc04f1264689b64dd8c09b2fe4c6daf07))

## [0.34.0](https://github.com/josephschmitt/monocle/compare/v0.33.0...v0.34.0) (2026-03-29)


### Features

* **release:** add prerelease beta release support for next branch ([f27ebaa](https://github.com/josephschmitt/monocle/commit/f27ebaa387ea529b4099afaa7a9680f48a56baf4))
* **tui:** add alt-based word navigation and deletion keybindings ([57bb0ab](https://github.com/josephschmitt/monocle/commit/57bb0abdf39f5f3bb026dab285df835e7bf686e1))


### Bug Fixes

* **release:** add prerelease versioning strategy for beta releases ([1093f32](https://github.com/josephschmitt/monocle/commit/1093f32a6039b5b0fce50490d17c4c7962712023))
* **release:** use numbered prerelease-type to ensure beta.0 suffix ([1ac4bde](https://github.com/josephschmitt/monocle/commit/1ac4bde3d7cc1b49db7c900cb323821635f0d7b4))
* **tui:** restore plan/artifact diff support for content item versions ([8e3b704](https://github.com/josephschmitt/monocle/commit/8e3b70400fd4546ef580789d337d55fb2cf7ede4))

## [0.33.0](https://github.com/josephschmitt/monocle/compare/v0.32.0...v0.33.0) (2026-03-28)


### Features

* **channel:** remove prescriptive workflow advice from MCP instructions ([c8025af](https://github.com/josephschmitt/monocle/commit/c8025afe6238444e6a88d82f84c979d98218671d))
* **core:** enforce request_changes reviews must include comments or body ([2b3d2eb](https://github.com/josephschmitt/monocle/commit/2b3d2eb526cb4f26fa7a65915a86eca0cd4213d7))
* **register:** overwrite existing config instead of no-oping ([8aa9257](https://github.com/josephschmitt/monocle/commit/8aa925775b92ab7a107f15910dfb79e9d9819a1e))


### Bug Fixes

* **channel:** make subagent tool restrictions more prominent in MCP descriptions ([5febd5f](https://github.com/josephschmitt/monocle/commit/5febd5f16ac756f2b0d6886174cae5fbae9240d4))

## [0.32.0](https://github.com/josephschmitt/monocle/compare/v0.31.0...v0.32.0) (2026-03-28)


### ⚠ BREAKING CHANGES

* **channel:** MCP tool names changed from submit_plan/submit_plan_and_wait to submit_for_review/submit_for_review_and_wait

### Features

* **opencode:** enable monocle MCP tools in plan mode during register ([b5a3fc1](https://github.com/josephschmitt/monocle/commit/b5a3fc1e07e354070fa8040e1129d78475674ce5))


### Code Refactoring

* **channel:** rename submit_plan tools to submit_for_review and simplify instructions ([680b529](https://github.com/josephschmitt/monocle/commit/680b529fd7f6b313a789d117b240e3b3a9864de7))

## [0.31.0](https://github.com/josephschmitt/monocle/compare/v0.30.0...v0.31.0) (2026-03-27)


### Features

* add MCP configs and slash commands for third-party agents ([029a14a](https://github.com/josephschmitt/monocle/commit/029a14a33f4841af930533bb28a7bcb60079fffa))
* **cli:** add multi-agent register command with interactive picker ([c95a9df](https://github.com/josephschmitt/monocle/commit/c95a9df29c0607b28376b9a4577b122263ab6a11))
* **core:** add queued feedback delivery with pull-based retrieval ([0699dff](https://github.com/josephschmitt/monocle/commit/0699dff7d4e790a0c52cc3791877ffbd61838cf8))
* **tui:** update splash screen for multi-agent support ([388e9ea](https://github.com/josephschmitt/monocle/commit/388e9ea4a35af682b37cbc71d2c4dc7ae8af9887))


### Bug Fixes

* **channel:** restrict submit_plan tools to top-level agent only ([76ba7f6](https://github.com/josephschmitt/monocle/commit/76ba7f6b3ba94ffbe6eeeb89a943edf1611770c6))
* **cli:** fix picker space key and use context-aware title ([eb0c2e1](https://github.com/josephschmitt/monocle/commit/eb0c2e15b66fc2f435b9e6adca246e9d86cc8b83))
* **cli:** scope HasConfig check to local/global and fix OpenCode path ([288b963](https://github.com/josephschmitt/monocle/commit/288b9639c492bbc2ff1fa8bbb154fd575b388234))
* **core:** reset lastKnownHead when re-enabling auto-advance base ref ([a52cc28](https://github.com/josephschmitt/monocle/commit/a52cc2890c9861763b1f79aaf8851a53f679930b))
* **tui:** prevent status bar cutoff in stacked layout when sidebar fills height ([ac3e018](https://github.com/josephschmitt/monocle/commit/ac3e01802b279135ec0722bf4529c9704bbcaecb))
* **tui:** show user-selected ref in status bar and picker, not diff parent ([712658c](https://github.com/josephschmitt/monocle/commit/712658ce4d0b3bb36a1f48357601a074e3914ac3))

## [0.27.0](https://github.com/josephschmitt/monocle/compare/v0.26.0...v0.27.0) (2026-03-26)


### Features

* **core:** auto-identify agent name via MCP channel ([3202d81](https://github.com/josephschmitt/monocle/commit/3202d81d702409ec0c032990ff3a12726ce3354b))
* **db:** add MONOCLE_DB env var to override database path ([170d1e1](https://github.com/josephschmitt/monocle/commit/170d1e13314d478e8508d2df098bd9c0262fba5d))
* **tui:** add clear review command to reset in-progress review ([5a7d2ab](https://github.com/josephschmitt/monocle/commit/5a7d2aba119e801b335dadd86f9b448c2da99da7))
* **tui:** add Ctrl+G to open external editor in comment and submit modals ([872cd90](https://github.com/josephschmitt/monocle/commit/872cd9055e9d1cbd6d9ff15f320713926545af42))
* **tui:** show agent name alongside connection status ([40e8f0c](https://github.com/josephschmitt/monocle/commit/40e8f0cbeefbdbdff2caed53aad040de5e42f041))


### Bug Fixes

* **tui:** keep agent name in default color next to connection status ([4d196ba](https://github.com/josephschmitt/monocle/commit/4d196baa63c22b4dcd91ff46baac6172297d59d7))

## [0.26.0](https://github.com/josephschmitt/monocle/compare/v0.25.0...v0.26.0) (2026-03-26)


### Features

* **tui:** remove clear-after-submit confirmation dialog ([7845287](https://github.com/josephschmitt/monocle/commit/7845287746955c4925206f08d298e51ff67a901a))


### Bug Fixes

* **channel:** prevent zombie MCP processes when Claude Code exits ([b5d6ddc](https://github.com/josephschmitt/monocle/commit/b5d6ddcf3094172eedd4d1bfd034956277d1c715))
* **core:** clear pending feedback after push delivery ([6277703](https://github.com/josephschmitt/monocle/commit/62777038e7fd36cf92fefb285809d1c8e8bd5877))
* **tui:** stop clearing files and base ref on submit ([98aa7a2](https://github.com/josephschmitt/monocle/commit/98aa7a2aabd6932c6b074e8ccbf3b74c3e2687e2))

## [0.25.0](https://github.com/josephschmitt/monocle/compare/v0.24.0...v0.25.0) (2026-03-25)


### Features

* **channel:** start MCP server without blocking on engine connection ([4666dd9](https://github.com/josephschmitt/monocle/commit/4666dd9a7ca768cfa8b0efb47daee62b6de67427))
* **plugin:** add slash command skills for sending plans to Monocle ([e41a86b](https://github.com/josephschmitt/monocle/commit/e41a86b356f636216476f02eec1f014e498178f7))


### Bug Fixes

* **channel:** simplify review approval text and propagate action in protocol ([9236c81](https://github.com/josephschmitt/monocle/commit/9236c81b218bd7852a092bf35948854731696a9b))
* **tui:** use visual-width truncation for non-wrap mode cutoff ([f494ef2](https://github.com/josephschmitt/monocle/commit/f494ef2dd86f5703e336a7955fc439e957e99ebe))

## [0.24.0](https://github.com/josephschmitt/monocle/compare/v0.23.0...v0.24.0) (2026-03-25)


### Features

* **tui:** add XML syntax highlighting for plist, xmp, and other XML dialects ([df93611](https://github.com/josephschmitt/monocle/commit/df936116a68066ea4a47e9055c91e4c882cd31d8))
* **tui:** support opening monocle in non-git directories ([cd4833f](https://github.com/josephschmitt/monocle/commit/cd4833f59da08677dc3092182558d90798e63561))

## [0.23.0](https://github.com/josephschmitt/monocle/compare/v0.22.0...v0.23.0) (2026-03-25)


### Features

* **tui:** add :mark-all-reviewed and :mark-all-unreviewed commands ([b7496d6](https://github.com/josephschmitt/monocle/commit/b7496d62312ca149b0becf912019fcf92af60ddb))
* **tui:** add configurable min_diff_width for side-by-side layout ([bb2f6c3](https://github.com/josephschmitt/monocle/commit/bb2f6c3d608cc7d8ea329d51463b30dab80cb29a))
* **tui:** add cursor navigation to comment editor modal ([2d2b577](https://github.com/josephschmitt/monocle/commit/2d2b577eb07567f409d3703c82690b76faa11b6b))
* **tui:** add emacs keybindings and smart home to comment editor ([0b1f474](https://github.com/josephschmitt/monocle/commit/0b1f474f9582a8945916960a5331426b9f455dd5))
* **tui:** add plugin install instructions to splash screen ([80ff8d3](https://github.com/josephschmitt/monocle/commit/80ff8d35182e90855a3a8f5bf0710b94994b3de2))
* **tui:** add suggested edits with GitHub-style suggestion blocks ([c68a2fb](https://github.com/josephschmitt/monocle/commit/c68a2fba8c43801331898f11bd14f3d811c01bd2))
* **tui:** cycle sidebar filter through all, unreviewed only, reviewed only ([910c22f](https://github.com/josephschmitt/monocle/commit/910c22f895f33d0f51e79e048b292bf271511a36))
* **tui:** enhance review marking with auto-advance, reset on submit, and filter ([54954c7](https://github.com/josephschmitt/monocle/commit/54954c7a29e88ee6a16a250423b5ecb420d5f9c2))


### Bug Fixes

* **channel:** remove ExitPlanMode references from MCP instructions ([454d548](https://github.com/josephschmitt/monocle/commit/454d5484bc5d7ef3aba1c499dc2df9a9a2d4477a))
* **channel:** simplify plan review instructions to avoid conflicting with native plan mode ([2c7e378](https://github.com/josephschmitt/monocle/commit/2c7e378ee3e660bb0f3c996439370eabf8d5ed2a))

## [0.22.0](https://github.com/josephschmitt/monocle/compare/v0.21.0...v0.22.0) (2026-03-25)


### Features

* **channel:** add file_path parameter to submit_plan tools ([a6646a6](https://github.com/josephschmitt/monocle/commit/a6646a6cded017026a0489c256c7142feb053882))
* **channel:** exit when Monocle not running, advance round on submit ([ed1a0e2](https://github.com/josephschmitt/monocle/commit/ed1a0e22a2be7968266b62d0e8b626bb82c742b6))
* **tui:** add focus mode with indicator badge and F keybind ([1b91078](https://github.com/josephschmitt/monocle/commit/1b910783fbbee019b5cd119b38f5cf978b8a50e4))


### Bug Fixes

* **adapters:** detect Claude Code plugin registration in NeedsRegister ([1539e79](https://github.com/josephschmitt/monocle/commit/1539e79de602bc541462bd284bcd2c4876ca5129))

## [0.21.0](https://github.com/josephschmitt/monocle/compare/v0.20.0...v0.21.0) (2026-03-24)


### Features

* add session continue/resume with --continue, --resume, and --session flags ([a9672e5](https://github.com/josephschmitt/monocle/commit/a9672e5a3ee1ae2fbf8a0d055408104735d3fbe7))

## [0.20.0](https://github.com/josephschmitt/monocle/compare/v0.19.0...v0.20.0) (2026-03-24)


### Features

* **tui:** add mouse mode support ([7469023](https://github.com/josephschmitt/monocle/commit/74690230dc1b354213efcb88b41b8337f1ef25fd))


### Bug Fixes

* **tui:** add mouseOriginY offset for Bubble Tea v2 alt-screen rendering ([4081a78](https://github.com/josephschmitt/monocle/commit/4081a787bd0b9e9ef4432d06159ee6c92624044e))
* **tui:** correct mouse click offset for multi-line comments ([0c70949](https://github.com/josephschmitt/monocle/commit/0c7094968cbbaab6e4a5fda0d0fd32d392b9ea38))
* **tui:** prevent terminal-level line wrapping in diff view ([4509651](https://github.com/josephschmitt/monocle/commit/45096513ed22ac6a6031a7e9113cb26729a0d3fd))

## [0.19.0](https://github.com/josephschmitt/monocle/compare/v0.18.1...v0.19.0) (2026-03-24)


### Features

* **tui:** add global keybindings, section navigation, sidebar toggle, and refresh ([c628c25](https://github.com/josephschmitt/monocle/commit/c628c25ef6c2618a3eabfcfa78f467d43600cf37))
* **tui:** add plan review mode for comfortable plan reviewing ([8635d37](https://github.com/josephschmitt/monocle/commit/8635d37813a48317812f3d38923054c2f51fabb8))
* **tui:** add vim-style horizontal scroll keybinds (0, ^, $) ([e5376bf](https://github.com/josephschmitt/monocle/commit/e5376bf7ccfe5fee97612db6f0611463815f5281))


### Bug Fixes

* **core:** clear content items on round advance ([3717b91](https://github.com/josephschmitt/monocle/commit/3717b91293a7290b02651c901607d5009fb4cc02))
* **tui:** clear diff view when file list becomes empty ([d317af7](https://github.com/josephschmitt/monocle/commit/d317af730245998e6a99a27d9f8bcbe394c31e48))
* **tui:** refresh diff view and content items during periodic tick ([8267128](https://github.com/josephschmitt/monocle/commit/8267128f6b50b1ff2da21474d6b493780ca5ecc9))
* **tui:** snap cursor to viewport edge when off-screen on j/k press ([177d9d0](https://github.com/josephschmitt/monocle/commit/177d9d0407d06e3e6e7b8e68ae026aa9e4113fcc))

## [0.18.1](https://github.com/josephschmitt/monocle/compare/v0.18.0...v0.18.1) (2026-03-24)


### Bug Fixes

* add required owner field to marketplace.json ([6124e1e](https://github.com/josephschmitt/monocle/commit/6124e1ed73036735408ec0cb6ba76b2ec1a5c3c7))
* **tui:** increase stacked layout sidebar minimum height and lower max cap ([25e2c00](https://github.com/josephschmitt/monocle/commit/25e2c008b2bf4b5c41708b7335f61694db95f1b1))

## [0.18.0](https://github.com/josephschmitt/monocle/compare/v0.17.0...v0.18.0) (2026-03-24)


### Features

* add additional files support for external file review ([ab3b257](https://github.com/josephschmitt/monocle/commit/ab3b257b7bbf5e3fa80b711ffc3d2c1f910790f6))
* **channel:** add Claude Code plugin for channel registration ([60522a5](https://github.com/josephschmitt/monocle/commit/60522a531a52b7c21ef20dd573241f480667532c))
* **channel:** add submit_plan_and_wait blocking tool for plan mode ([8e4ffc9](https://github.com/josephschmitt/monocle/commit/8e4ffc973fd7d74ccbe9ebca76ec75773d6a8e5f))


### Bug Fixes

* **tui:** sync sidebar cursor with diff viewer and enable content review toggle ([af7ae5a](https://github.com/josephschmitt/monocle/commit/af7ae5a8dc84e050d7aca4d92adec0072f5c1f37))

## [0.17.0](https://github.com/josephschmitt/monocle/compare/v0.16.0...v0.17.0) (2026-03-23)


### Features

* **tui:** add file view mode to diff style cycle ([6dca8c3](https://github.com/josephschmitt/monocle/commit/6dca8c3706d1be6b7a3b35e050bf89993e86d49e))


### Bug Fixes

* **channel:** stop redundant get_feedback call after feedback submission ([8dcdf93](https://github.com/josephschmitt/monocle/commit/8dcdf93e61fccdb71fd4daec15351159734287d3))
* **core:** include untracked files in diff view ([592f852](https://github.com/josephschmitt/monocle/commit/592f85272500bed748be551a5c63672fba66c9c0))
* **tui:** preserve markdown styling when word-wrap is toggled on ([ee7eb05](https://github.com/josephschmitt/monocle/commit/ee7eb059f3d09ea21be73b24f3541602f94f856c))
* **tui:** update splash screen to use correct Claude Code channels flag ([85a5e19](https://github.com/josephschmitt/monocle/commit/85a5e19e5ac28b8a4a6d1491f0b6ec50f75a4734))

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
