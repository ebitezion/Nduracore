/**
 * Custom sidebar items generator that injects “top” and “bottom” section labels
 * around individual docs based on front-matter flags.
 *
 * To use, add these to the **YAML front-matter** of your Markdown file (no quotes needed):
 *
 * ```md
 * ---
 * title: Changelog
 * sectionTopLabel: Introduction     # injects “Introduction” above this doc
 * sectionBottomLabel: Versions      # injects “Versions” below this doc
 * ---
 * ```
 *
 * @param args
 *   - defaultSidebarItemsGenerator: function to produce the normal sidebar items
 *   - docs: array of all doc metadata (including your frontMatter)
 * @returns
 *   A new array of sidebar items, with `<div class="sidebar-section-label">…</div>`
 *   injected before/after any doc that defines `sectionTopLabel` or
 *   `sectionBottomLabel` in its front-matter.
 */
export async function sidebarLabelGenerator(args) {
  const defaultItems = await args.defaultSidebarItemsGenerator!(args);
  return defaultItems.flatMap((item) => {
    // only modify real docs (skip categories, html, etc.)
    if (item.type === "doc") {
      const docMeta = args.docs.find((d) => d.id === item.id);
      const { sectionTopLabel, sectionBottomLabel } =
        docMeta?.frontMatter || {};
      const out = [];

      // If the front-matter had a sectionTopLabel, inject it above
      if (sectionTopLabel) {
        out.push({
          type: "html",
          value: `<div class="sidebarSectionLabel">${sectionTopLabel}</div>`,
          defaultStyle: true,
        });
      }

      out.push(item);

      // If the front-matter had a sectionBottomLabel, inject it below
      if (sectionBottomLabel) {
        out.push({
          type: "html",
          value: `<div class="sidebarSectionLabel">${sectionBottomLabel}</div>`,
          defaultStyle: true,
        });
      }

      return out;
    }

    return [item];
  });
}
