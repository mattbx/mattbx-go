package ui

// Shared Tailwind utility clusters for templates. Keep as utility strings
// (listed in @source); don't invent semantic class names in main.css.

const (
	classPageHead = "mb-16"
	classEyebrow  = "font-mono text-meta tracking-[0.1em] uppercase text-dim m-0 mb-2"
	classPageTitle = "font-ui text-hero font-semibold tracking-tight m-0 text-fg"
	classPageTitleSm = "font-ui text-title font-semibold tracking-tight m-0 text-fg"
	classPageLede    = "mt-2"
	classSectionHead = "flex flex-wrap items-baseline justify-between gap-3 mt-16 mb-3"
	classEmpty       = "border border-dashed border-line rounded-[3px] p-10 text-center text-dim font-ui text-meta"
	classBadge       = "inline-block font-mono text-micro tracking-wider uppercase px-[0.5em] py-[0.15em] rounded-[3px] border border-current leading-normal"
	classNotice      = "font-ui text-meta px-4 py-3 rounded-[3px] border-l-2 border-danger bg-surface text-fg mb-4"
	classSkipLink    = "sr-only focus:not-sr-only focus:absolute focus:left-3 focus:top-2 focus:z-10 focus:bg-fg focus:text-page focus:px-3 focus:py-2 focus:rounded-[3px]"
)
