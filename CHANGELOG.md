# Changelog

## [0.46.0-beta.0](https://github.com/josephschmitt/monocle/compare/v0.45.0...v0.46.0-beta.0) (2026-04-18)


### Features

* **core:** add review tracking with snapshot-based change detection ([d634347](https://github.com/josephschmitt/monocle/commit/d634347d5b35cde37e8c294e19aa834b6ee04b28))
* **db:** add review snapshot schema and queries ([15e18c1](https://github.com/josephschmitt/monocle/commit/15e18c1dc3a7439357815a8e6c0ec0cdf59bdd54))
* **desktop:** {/} jump between comments in diff view ([83a0997](https://github.com/josephschmitt/monocle/commit/83a0997f01ce1d344f878f94fedc99b4d0bcf03e))
* **desktop:** add additional files via sidebar and palette ([a57dc2a](https://github.com/josephschmitt/monocle/commit/a57dc2a177d3dbbd18d2b376e2742842e1941a0d))
* **desktop:** add artifact version picker and version-diff mode ([4e5dc8f](https://github.com/josephschmitt/monocle/commit/4e5dc8faeccd3106a9478913233ade588974a08e))
* **desktop:** add cancel pause command to palette ([24472fb](https://github.com/josephschmitt/monocle/commit/24472fb87c25825603e6fced66a86b62b469bf4f))
* **desktop:** add comment editor, review submission, and review actions ([7c23be6](https://github.com/josephschmitt/monocle/commit/7c23be6f583f4ea196ab2d2996ab35635f0341f5))
* **desktop:** add connection info, history, and base ref picker dialogs ([05a33b2](https://github.com/josephschmitt/monocle/commit/05a33b2c1d0378259ac14b15fe3a51166ed01a64))
* **desktop:** add custom app icon with monocle logotype ([275f408](https://github.com/josephschmitt/monocle/commit/275f40869c9443fb002b010ed13a7759056b80b1))
* **desktop:** add diff view keyboard navigation and fix shift key matching ([0f07a58](https://github.com/josephschmitt/monocle/commit/0f07a58ad30219d30031c3a03db36be218fd888e))
* **desktop:** add diff view with react-diff-view and content viewer ([9a6dbbf](https://github.com/josephschmitt/monocle/commit/9a6dbbfae56dc4d62296d5f3a789b817fd3668a8))
* **desktop:** add Enter/dir toggle, tree collapse/expand, and Ctrl+Y copy ([c671c99](https://github.com/josephschmitt/monocle/commit/c671c995c200a8155087accf9a54c33b35133766))
* **desktop:** add File &gt; Open Project menu for switching projects ([1d7ddb7](https://github.com/josephschmitt/monocle/commit/1d7ddb71f99e77fe42268b91269859ca163215a0))
* **desktop:** add full splash screen matching TUI empty state ([9455b59](https://github.com/josephschmitt/monocle/commit/9455b59af3b8162eb83f8c4d633ce4394aa8a1ab))
* **desktop:** add help dialog, command palette, and remaining keyboard shortcuts ([34af0f6](https://github.com/josephschmitt/monocle/commit/34af0f6a17bc8d7aba45bd84f32e48099e061351))
* **desktop:** add layout cycling between horizontal and stacked ([4a5c0ed](https://github.com/josephschmitt/monocle/commit/4a5c0ede1217cea3266f244070a55a310968f76d))
* **desktop:** add layout shell, sidebar, keyboard nav, and API layer ([0f05fb2](https://github.com/josephschmitt/monocle/commit/0f05fb2abaa340bd24bfa86bb7ed1dc129341665))
* **desktop:** add macOS-native frameless window with toolbar and refined sidebar ([914fcf3](https://github.com/josephschmitt/monocle/commit/914fcf320ecd262a6d4af30379a176c0dfa006e0))
* **desktop:** add missing keybindings for TUI feature parity ([21f74d6](https://github.com/josephschmitt/monocle/commit/21f74d6118517a579c777d6f8d443d5b6788477f))
* **desktop:** add Nerd Font file icons in sidebar ([f8812f8](https://github.com/josephschmitt/monocle/commit/f8812f8165b98ca8dacc4332440ee8a8cb1df290))
* **desktop:** add project picker for directory selection ([6f6d5f1](https://github.com/josephschmitt/monocle/commit/6f6d5f11e8f43e6cfc6adee6952a38a25d833e73))
* **desktop:** add project switcher dropdown to toolbar ([5c97e49](https://github.com/josephschmitt/monocle/commit/5c97e493fe137a7813fb6211dc48ae1cdd4190b2))
* **desktop:** add session picker on startup ([d003c34](https://github.com/josephschmitt/monocle/commit/d003c34a29b0f0d38088280f533b5a08806f64f0))
* **desktop:** add Settings dialog ([0e3b4fe](https://github.com/josephschmitt/monocle/commit/0e3b4fe3514dee73b6fd2c7e0095b43e0eadc154))
* **desktop:** add submit! auto-submit palette command ([c098d5e](https://github.com/josephschmitt/monocle/commit/c098d5e651908361585e00a7fe68981e245ffb33))
* **desktop:** add syntax highlighting to diff view ([4f10146](https://github.com/josephschmitt/monocle/commit/4f101460ec7103969859e6dcab2e5c63d41fe949))
* **desktop:** add toast notification system ([c5b7b2b](https://github.com/josephschmitt/monocle/commit/c5b7b2b553f18c17841e6e033c3d9f9b7ac28ef7))
* **desktop:** add visual selection, click-to-focus, and drag-to-select ([3faa806](https://github.com/josephschmitt/monocle/commit/3faa8066b7f01a25a8290b2eddb0e6dc6e3e2b57))
* **desktop:** add website-matched typography and visual polish ([a6f7c65](https://github.com/josephschmitt/monocle/commit/a6f7c656e5713127e5af80410b5030af606df43f))
* **desktop:** auto-advance to next unreviewed after marking reviewed ([f62a7d7](https://github.com/josephschmitt/monocle/commit/f62a7d79dafe11f7a2aded81e46dcd0911f60ccc))
* **desktop:** auto-select plans when received and fix null hunks crash ([858382e](https://github.com/josephschmitt/monocle/commit/858382e8d3bd600911a19a965d636fd2f2b59234))
* **desktop:** bring status bar and toolbar to TUI parity ([4572993](https://github.com/josephschmitt/monocle/commit/4572993de3b2826601a36d7b79f54205f2a3f3d0))
* **desktop:** collapse single-child directory chains in sidebar tree ([a5f309b](https://github.com/josephschmitt/monocle/commit/a5f309b03c8bfa90f73f841bd0c8e3d97a05b59a))
* **desktop:** confirm destructive review commands ([823c768](https://github.com/josephschmitt/monocle/commit/823c768d62dcd80e55d2501b0b998f52b8235c71))
* **desktop:** contextual empty state when review tracking is active ([afe030e](https://github.com/josephschmitt/monocle/commit/afe030e6c34e4aad7ee2ea08bdec553f549b310d))
* **desktop:** honor config.keybindings overrides ([da9d2a3](https://github.com/josephschmitt/monocle/commit/da9d2a34b5596ba1868f2e30ad81ad679d5d50f7))
* **desktop:** improve base ref picker and show current ref in status bar ([0c3fbd3](https://github.com/josephschmitt/monocle/commit/0c3fbd3961de029a5e21f61c8bc4a3526895f0fe))
* **desktop:** make comments navigable with j/k cursor movement ([3fd2cd7](https://github.com/josephschmitt/monocle/commit/3fd2cd75c09a6edf2e0d73fd296f558eed0649ec))
* **desktop:** match TUI connection indicator states in toolbar ([f64894e](https://github.com/josephschmitt/monocle/commit/f64894e66202460b6c6df777783db70652758d64))
* **desktop:** move logotype next to traffic lights using JetBrains Mono ([5a63e81](https://github.com/josephschmitt/monocle/commit/5a63e813e4f3bce0a1a1637809b077852f612297))
* **desktop:** offer Claude MCP registration on first run ([e5ade22](https://github.com/josephschmitt/monocle/commit/e5ade22c4cd6bf02649c348b1ff4379025269c60))
* **desktop:** open comment and review bodies in external editor ([2eaf50c](https://github.com/josephschmitt/monocle/commit/2eaf50c74181a1544183c1adc758e9b3872f9a12))
* **desktop:** press c on focused comment to edit it ([575e356](https://github.com/josephschmitt/monocle/commit/575e3566540f71700fe723b3f25babcb0e4945b8))
* **desktop:** publish Homebrew cask on release ([7f124a4](https://github.com/josephschmitt/monocle/commit/7f124a41b866b42a0af1f9f336838f2e8ac9526a))
* **desktop:** render file contents in non-git directory mode ([1d7222d](https://github.com/josephschmitt/monocle/commit/1d7222de01d12659a4e11148b39ec7d9acd7d244))
* **desktop:** render file-level comments at top of diff view ([2e9b201](https://github.com/josephschmitt/monocle/commit/2e9b201fced3e313a2888026bc7049580e0ad158))
* **desktop:** render plans and files through DiffView with full navigation ([444832b](https://github.com/josephschmitt/monocle/commit/444832b0795b43be95fd34a6e493861616fd1c49))
* **desktop:** render suggestion comments as inline diffs ([56c847a](https://github.com/josephschmitt/monocle/commit/56c847a5d1e2ff4debd2259c41e0786bfb010525))
* **desktop:** replace border focus indicators with accent bars and kill native outlines ([590073d](https://github.com/josephschmitt/monocle/commit/590073dd3cdb14010379e5bf31faac2cd2894fee))
* **desktop:** reposition macOS traffic lights to align with toolbar ([12f96c6](https://github.com/josephschmitt/monocle/commit/12f96c683f2b214f646cd5ff52ba93dc03f18af2))
* **desktop:** show binary file indicator instead of rendering garbled content ([529087f](https://github.com/josephschmitt/monocle/commit/529087fa820e9732c57422ca23bf8d168f469a0a))
* **desktop:** show empty circle for unreviewed files in sidebar ([f24d73f](https://github.com/josephschmitt/monocle/commit/f24d73f09b1cdfd1c927e9c2d990a04930c49455))
* **desktop:** surface directory mode notice on startup ([dd44e3b](https://github.com/josephschmitt/monocle/commit/dd44e3b2b2f4460a23f3f95d11570e0d07675e94))
* **desktop:** surface review snapshots in base ref picker and status bar ([70313c7](https://github.com/josephschmitt/monocle/commit/70313c78c62f605401b0f812cfa87806c286719b))
* **desktop:** switch to Shiki for diff syntax highlighting ([8fcea08](https://github.com/josephschmitt/monocle/commit/8fcea08629b7d0854a128316af63b8380d7fc7cb))
* **desktop:** update help dialog with all TUI-parity keybindings ([a52d4e0](https://github.com/josephschmitt/monocle/commit/a52d4e0803e008865f85b07a107b5a774b2becb5))
* scaffold Wails desktop app with React frontend ([8a2fd5b](https://github.com/josephschmitt/monocle/commit/8a2fd5b446fe22522691946ba893f1ed23b84af8))
* **tui:** render review indicators, snapshot ref picker, and stable sidebar ([5382b38](https://github.com/josephschmitt/monocle/commit/5382b383a30b4c140160a2a62a65afbae0752133))


### Bug Fixes

* **ci:** bundle MCP channel before desktop builds ([2ef082f](https://github.com/josephschmitt/monocle/commit/2ef082f3b9760b45ced46d837f960de7e24cc3ea))
* **ci:** remove MCP channel bundle step from desktop workflows ([b85e00b](https://github.com/josephschmitt/monocle/commit/b85e00b522e3e26f800e70a260960a71a3cf67ea))
* **ci:** replace macos-13 matrix with cross-compilation on macos-latest ([9ef8952](https://github.com/josephschmitt/monocle/commit/9ef8952edbed7d7b7adabc5d25b12550f6f40632))
* **ci:** simplify desktop release to arm64-only build ([dc63d53](https://github.com/josephschmitt/monocle/commit/dc63d53e0ead0de41b407f75099daed93517b42d))
* **desktop:** add file type icon to binary file indicator ([18440d0](https://github.com/josephschmitt/monocle/commit/18440d0c8878bc06732260d945bc11bc32fce7fb))
* **desktop:** align git status left and review checkmark right in sidebar ([eab4afa](https://github.com/josephschmitt/monocle/commit/eab4afafb57e254a981d48a2fd45ce3b31d4c33b))
* **desktop:** cursor movement now triggers file selection ([3243849](https://github.com/josephschmitt/monocle/commit/32438492aca7b8e015f50da904a8e1e0b479eccd))
* **desktop:** dim selection highlights in unfocused panes ([aa49fa8](https://github.com/josephschmitt/monocle/commit/aa49fa85a980e6c8d9baade0fc1047bff42582e2))
* **desktop:** downgrade refractor to v4 for react-diff-view compat ([91a26fd](https://github.com/josephschmitt/monocle/commit/91a26fdcc9a648182bfde4f66812d941ae83f0dd))
* **desktop:** fix base ref picker footer overlap and scroll overflow ([e6fa7b9](https://github.com/josephschmitt/monocle/commit/e6fa7b93790cf395b07b8881f432a5059d71f38d))
* **desktop:** fix connection indicator by tracking queue mode ([de97749](https://github.com/josephschmitt/monocle/commit/de97749bc060fab24edc5db2f55652bf3a03e6b0))
* **desktop:** isolate dev DB path and guard nil database in SelectProject ([b90639a](https://github.com/josephschmitt/monocle/commit/b90639a98c05f9c49ea52dc61bceae1fec999441))
* **desktop:** load config on startup to apply user preferences ([b671198](https://github.com/josephschmitt/monocle/commit/b6711988ae072b5827470afc3406d87347407232))
* **desktop:** load react-diff-view CSS before Catppuccin overrides ([f445959](https://github.com/josephschmitt/monocle/commit/f4459593f5103129563bf83f9ad5b4cdde757305))
* **desktop:** make sidebar full-height and add status bar padding ([b7dd885](https://github.com/josephschmitt/monocle/commit/b7dd885d98e5be7990fe6cf893026d3d36b6e5ac))
* **desktop:** move git status badge to left of file icon in sidebar ([b69595d](https://github.com/josephschmitt/monocle/commit/b69595d96b59826c49749e329462964c5ea01e3d))
* **desktop:** pass structural tokens so renderToken fires for Shiki ([8c401b5](https://github.com/josephschmitt/monocle/commit/8c401b5d285235a50313bd826c8e6c5a5f328c68))
* **desktop:** pre-fill suggestion block with selected line content ([f5964ff](https://github.com/josephschmitt/monocle/commit/f5964ff9641d343f1fd01cac3c931c2eb3eb5559))
* **desktop:** prevent duplicate comment widgets on modified lines ([7c68d30](https://github.com/josephschmitt/monocle/commit/7c68d301e383846b07713a789773ae7a293680d0))
* **desktop:** reduce sidebar item font size from text-sm to text-xs ([4dc831c](https://github.com/josephschmitt/monocle/commit/4dc831c1c7328440e8c839287ffef6a51a22a708))
* **desktop:** refresh changed files after base ref selection ([d725a21](https://github.com/josephschmitt/monocle/commit/d725a219623513d28c8239e8e8dffefb5019f4d7))
* **desktop:** refresh diff view after submit and clear ([b11bf86](https://github.com/josephschmitt/monocle/commit/b11bf86ef7b87a7284f47494ffc3fa28db8fedd1))
* **desktop:** remove fixed width on sidebar icons to prevent overlap ([53cb9bd](https://github.com/josephschmitt/monocle/commit/53cb9bd237898614a7f99ac860dbc7e40e5c080a))
* **desktop:** rename app bundle to Monocle.app and fix icon transparency ([d8d9cdb](https://github.com/josephschmitt/monocle/commit/d8d9cdbd7f952b17d35ce6b7dec70f4102d24b7e))
* **desktop:** return value from SelectProject for Wails binding compat ([9f77d06](https://github.com/josephschmitt/monocle/commit/9f77d0680d986eb482f8882b26ed0e4985bb81dd))
* **desktop:** scope dev-mode watcher to desktop-relevant directories ([4358954](https://github.com/josephschmitt/monocle/commit/4358954e481268e76bacb4c8cd7648b6963ae5bb))
* **desktop:** show ⌘ instead of Ctrl on Mac for keyboard hints ([72784bc](https://github.com/josephschmitt/monocle/commit/72784bcc1e351ba657ab8135205e2e25392391f5))
* **desktop:** show logotype and Focus Mode badge in toolbar when sidebar hidden ([96c5fc4](https://github.com/josephschmitt/monocle/commit/96c5fc471814164ee206ca3741b4c62affcf1a85))
* **desktop:** simplify Makefile build targets ([cc9fc02](https://github.com/josephschmitt/monocle/commit/cc9fc027869e09bafe4c141d66f3ec8f2ac135b9))
* **desktop:** skip folders when navigating files with [/] ([9146080](https://github.com/josephschmitt/monocle/commit/914608082440c95563b9cd89341883ff15f4da21))
* **desktop:** sort sidebar tree directories-first then alphabetical ([4b2abda](https://github.com/josephschmitt/monocle/commit/4b2abdaef1d56b3e09575bd62ef7c23304579b11))
* **desktop:** suppress file active highlight when cursor is on a folder ([549cbc7](https://github.com/josephschmitt/monocle/commit/549cbc70c4006e5fd9ce1f3498b0861fdd636f6f))
* **desktop:** surface project selection errors in picker ([e0a35dc](https://github.com/josephschmitt/monocle/commit/e0a35dcf78401aff6d9409e773dcc05a1bc16c35))
* **desktop:** update empty state to match TUI splash screen ([9cb5792](https://github.com/josephschmitt/monocle/commit/9cb57923bdc2c9a34e3dd7971783145f260c05c4))
* **desktop:** use bundled PureNerdFont for file icons ([a09da9d](https://github.com/josephschmitt/monocle/commit/a09da9dd8c594a35d1576d32c4bdc6ce9e4ae5b2))
* **desktop:** use monospace font in comment textareas and fix dev setup ([3165516](https://github.com/josephschmitt/monocle/commit/3165516a31bdc97835125153c1b52a584308a574))
* **desktop:** use Nerd Font family for sidebar file icons ([5471ba6](https://github.com/josephschmitt/monocle/commit/5471ba624ea4add27420758263c4366e2a769e24))
* **desktop:** use nerd font folder icon matching TUI sidebar ([dd10742](https://github.com/josephschmitt/monocle/commit/dd107429df01d089c5b4d21a78edbe245003b26a))
* **desktop:** use same highlight color for cursor and active sidebar items ([c9a42c6](https://github.com/josephschmitt/monocle/commit/c9a42c6198e08dc0e2075d4b85e7a5a0957f64d1))
* **desktop:** use solid fill color for focused comment highlight ([f7f121c](https://github.com/josephschmitt/monocle/commit/f7f121ca017f3e3e55ce3c0b4e39d8d4f024fdaf))
* **desktop:** use vibrant color for focused comment highlight ([99a4aab](https://github.com/josephschmitt/monocle/commit/99a4aabd3422e53fb05f7b7719fdde2b414fbbb2))
* **desktop:** widen comment and submit modals to max-w-2xl ([65ca60d](https://github.com/josephschmitt/monocle/commit/65ca60d740f3903a589866a590bcea91c35c05c2))
* **desktop:** wrap around when jumping between comments with {/} ([f6e56db](https://github.com/josephschmitt/monocle/commit/f6e56dbe49a7114bf53682dd2665568fea4bbd8d))
* make vet depend on frontend-dist so go vet works without a prior desktop build ([3f589db](https://github.com/josephschmitt/monocle/commit/3f589db7c14fc5dea014507cd002865ceea19614))

## [0.45.0](https://github.com/josephschmitt/monocle/compare/v0.44.0...v0.45.0) (2026-04-17)


### Features

* auto-review ExitPlanMode via Claude Code hooks ([d351f42](https://github.com/josephschmitt/monocle/commit/d351f4277dd759e4751326201d3092592c53d248))
* **hooks:** add per-turn review gate via PostToolUse + Stop hooks ([a62ee4a](https://github.com/josephschmitt/monocle/commit/a62ee4a9ee0bdb8412b0d07bb7ce848a1876f900))
* **register:** rebuild register/unregister as a themed TUI wizard ([ea6a07a](https://github.com/josephschmitt/monocle/commit/ea6a07afd4fd3780278e0b90bea0f36b9d64d121))


### Bug Fixes

* **hooks:** use absolute/repo-relative binary path + planFilePath as stable id ([40b13dc](https://github.com/josephschmitt/monocle/commit/40b13dc6993ddfa921418868b91a352d5caaca2b))

## [0.44.0](https://github.com/josephschmitt/monocle/compare/v0.43.0...v0.44.0) (2026-04-08)


### Features

* add --workdir/-C flag to override working directory for socket pairing ([33ecddb](https://github.com/josephschmitt/monocle/commit/33ecddbbc1a981e467703e102987828193a8274e))

## [0.43.0](https://github.com/josephschmitt/monocle/compare/v0.42.0...v0.43.0) (2026-04-08)


### Features

* **tui:** add g/G keybindings to ref picker for top/bottom jump ([02518c0](https://github.com/josephschmitt/monocle/commit/02518c0375d2ee0d0a354172cd3a900919830d29))
* **tui:** detect binary files and show indicator in diff view ([d7f343d](https://github.com/josephschmitt/monocle/commit/d7f343daef9b0f2475cf654d6a50d02c70a95bc1))

## [0.42.0](https://github.com/josephschmitt/monocle/compare/v0.41.1...v0.42.0) (2026-04-07)


### ⚠ BREAKING CHANGES

* serve-mcp-channel no longer starts a channel server. Use serve-mcp --experimental-channels instead (coming next).

### Features

* **adapters:** add embedded command definitions and install functions ([6828946](https://github.com/josephschmitt/monocle/commit/6828946f57792a7b89f9f0961078e0f0dcc3369a))
* **adapters:** add integration mode selection for Claude registration ([78a9af7](https://github.com/josephschmitt/monocle/commit/78a9af711d39d2b22904f3f2591f0536ad31628a))
* **adapters:** add MCP mode support to OpenCode and Gemini adapters ([addabb7](https://github.com/josephschmitt/monocle/commit/addabb76f328e97a4e69f6fec5e8f40de1831dcb))
* **adapters:** configure MCP server for all agents in MCP mode ([0e234b0](https://github.com/josephschmitt/monocle/commit/0e234b0c9bf3e7e8b7be7c1c9f322a8b632db101))
* **adapters:** install commands in MCP mode for Claude registration ([cb0c95c](https://github.com/josephschmitt/monocle/commit/cb0c95cae2e573d0634d8acab44a0693f3421e81))
* **adapters:** update registration to use serve-mcp ([3eb2569](https://github.com/josephschmitt/monocle/commit/3eb2569d36991285de02f56b3d5ba03aa2b251ad))
* **mcp:** add experimental channel support for push notifications ([eb6af21](https://github.com/josephschmitt/monocle/commit/eb6af21f5a487c989426c72324fcb3b0dcd0f5f1))
* **mcp:** add file_path param to send_artifact tool ([23d2043](https://github.com/josephschmitt/monocle/commit/23d2043f815c6c1dc3d793c693a853645d513c7d))
* **mcp:** add Go MCP server with review tools ([755dc24](https://github.com/josephschmitt/monocle/commit/755dc249318028ba56260f7e030cc2ee93bf6884))
* **mcp:** resolve engine socket from MCP client roots ([0092d78](https://github.com/josephschmitt/monocle/commit/0092d780301b08b5f5d3065b43d260e145abd366))


### Bug Fixes

* **adapters:** expand config paths on hover in agent picker ([63be593](https://github.com/josephschmitt/monocle/commit/63be59361c58c93cc996610014af9877b59400f1))
* **adapters:** fall back to latest release when exact skills version not found ([d1dad74](https://github.com/josephschmitt/monocle/commit/d1dad74859f11d9490f2ac53003ec3e9ece8dde9))
* **adapters:** set integration mode before picker shows config paths ([4488f68](https://github.com/josephschmitt/monocle/commit/4488f68de7eef32ad579f2b0cf14ebdb486c8c34))
* **adapters:** truncate long config paths in agent picker ([d30683f](https://github.com/josephschmitt/monocle/commit/d30683f4f465504ea1ce69eed7924b683eae39fd))
* **mcp:** make optional tool params non-required in schema ([9b89711](https://github.com/josephschmitt/monocle/commit/9b897112bf00735a621ef4f07fdc26e760ccb284))
* pass version ldflags in make install ([c5aaceb](https://github.com/josephschmitt/monocle/commit/c5aaceb10e61f047f564b5312652070546df532b))


### Code Refactoring

* remove TypeScript MCP channel server and JS build chain ([494f114](https://github.com/josephschmitt/monocle/commit/494f1149d7c82f9eb7cdb5e0f218862f1ffe7d32))

## [0.41.1](https://github.com/josephschmitt/monocle/compare/v0.41.0...v0.41.1) (2026-04-02)


### Bug Fixes

* **db:** scope plan version history to session ([c1b096c](https://github.com/josephschmitt/monocle/commit/c1b096c56b4784a9b37a7009006de199b254dfdf))
* **tui:** auto-select latest content item in viewer when artifact arrives ([7496e3b](https://github.com/josephschmitt/monocle/commit/7496e3b6aaff74a009292bed6ad2fc0e0437c29f))
* **tui:** correct sidebar height calculation for section headers ([62a9e4b](https://github.com/josephschmitt/monocle/commit/62a9e4bec548fa73fc59a2881d1c92e160b41fb4))

## [0.41.0](https://github.com/josephschmitt/monocle/compare/v0.40.0...v0.41.0) (2026-04-02)


### Features

* **tui:** add artifact version history browser ([a7154c0](https://github.com/josephschmitt/monocle/commit/a7154c045123c41c381b5bfcd21391f3618fd45e))
* **tui:** add B keybind for artifact versions and new base commands ([6430d6b](https://github.com/josephschmitt/monocle/commit/6430d6b57620485090a369b1711cf6e8f8ef620c))
* **tui:** add comment_expand and comment_expand_delay config settings ([272854a](https://github.com/josephschmitt/monocle/commit/272854a68c0958aaa74605651cc63984fcde125d))
* **tui:** add diff preview for suggestion comments in expanded state ([62f08ca](https://github.com/josephschmitt/monocle/commit/62f08caabe42d1dd98ca7e3ad56fc4ba32ddad2d))
* **tui:** auto-select saved comment in diff view ([d6519ea](https://github.com/josephschmitt/monocle/commit/d6519ea57ed4fe56d08bc75857da917ee9694626))


### Bug Fixes

* **tui:** respect wrap toggle and fix syntax highlighting in suggestion diffs ([b4acec4](https://github.com/josephschmitt/monocle/commit/b4acec43cd5df7249ed3f974200b985e645e5c6f))

## [0.40.0](https://github.com/josephschmitt/monocle/compare/v0.39.0...v0.40.0) (2026-03-31)


### Features

* **tui:** add {/} comment jumping with wrap-around in viewer pane ([cc38630](https://github.com/josephschmitt/monocle/commit/cc3863000f7c3ad6f9f76345a860c699e43e0e55))
* **tui:** add space to toggle comment expand and increase delay to 2s ([f7974cd](https://github.com/josephschmitt/monocle/commit/f7974cd10f052f28b9268b53550bc7dbebac6281))
* **tui:** expand inline comments on hover after a short delay ([450c6bf](https://github.com/josephschmitt/monocle/commit/450c6bf091ee05a625c8ae5c12261c8725d9cd6d))
* **tui:** use thick border for expanded comments and add syntax highlighting ([a364af6](https://github.com/josephschmitt/monocle/commit/a364af6f6277876ff1fc7b2b03bae81a6e419c2e))
* **website:** add interactive asciinema demos to feature cards ([0b23320](https://github.com/josephschmitt/monocle/commit/0b233200ecca1cd1a29b75a4f0ec0e76f50f6ac5))


### Bug Fixes

* persist comment type changes when editing a drafted comment ([f38ce31](https://github.com/josephschmitt/monocle/commit/f38ce31533c83c358f250a9aa69855ae60de4ea7))

## [0.39.0](https://github.com/josephschmitt/monocle/compare/v0.38.0...v0.39.0) (2026-03-30)


### Features

* **skills:** add get-feedback-wait skill for blocking review feedback ([4c9ccee](https://github.com/josephschmitt/monocle/commit/4c9cceebee43aac76905710d1069b4a53c8dadaa))

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
