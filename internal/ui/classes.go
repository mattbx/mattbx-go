package ui

// Shared Tailwind utility clusters for templates. Keep as utility strings
// (listed in @source); don't invent semantic class names in main.css.
//
// Type is one role class per element (type-page-title, type-meta, ...),
// defined in tailwind/typography.css. Don't rebuild type from font-*,
// text-<size>, tracking-*, leading-* or case utilities here: add or edit a role
// there instead. Colour, margin and layout stay as plain utilities.

const (
	classPageHead    = "mb-10"
	classEyebrow     = "type-eyebrow text-dim m-0 mb-2"
	classPageTitle   = "type-page-title m-0 text-fg"
	classPageTitleSm = "type-page-title m-0 text-fg"
	classPageLede    = "mt-2 type-lede text-dim"
	classSectionHead = "flex flex-wrap items-baseline justify-between gap-3 mt-10 mb-3"
	classEmpty       = "border border-dashed border-line rounded-[3px] p-10 text-center text-dim type-meta"
	classBadge       = "inline-block type-badge px-[0.5em] py-[0.15em] rounded-[3px] border border-current"
	classNotice      = "type-notice px-4 py-3 rounded-[3px] border-l-2 border-danger bg-surface text-fg mb-4"
	classSkipLink    = "sr-only focus:not-sr-only focus:absolute focus:left-3 focus:top-2 focus:z-10 focus:bg-fg focus:text-page focus:px-3 focus:py-2 focus:rounded-[3px]"
)
