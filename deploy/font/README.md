将用于 OG 图片渲染的字体放到本目录，以避免在服务器构建镜像时在线下载（速度慢/不稳定）。

## 需要的字体类型
- 支持 `.otf` / `.ttf`（`resvg` 在 musl 环境下无法加载 `.woff2`）

## 推荐字体（与默认 Dockerfile 逻辑一致）
- **Noto Serif SC**：建议至少包含 `NotoSerifSC-Regular.otf`、`NotoSerifSC-Bold.otf`
- **Google Sans Code**：建议放入 `GoogleSansCode` 的 `.ttf`（可变字体或静态 TTF 均可）

## 构建时如何生效
`deploy/docker/web.Dockerfile` 会执行：
- `COPY deploy/font/ /usr/share/fonts/og/`
- `fc-cache -f`

