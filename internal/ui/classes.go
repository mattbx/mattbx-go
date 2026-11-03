package ui

// Shared Tailwind utility clusters for templates. Keep as utility strings
// (listed in @source); don't invent semantic class names in main.css.

const (
	classPageHead    = "mb-10"
	classEyebrow     = "font-mono text-meta tracking-normal uppercase text-dim m-0 mb-2"
	classPageTitle   = "font-ui text-section font-semibold tracking-tight m-0 text-fg"
	classPageTitleSm = "font-ui text-section font-semibold tracking-tight m-0 text-fg"
	classPageLede    = "mt-2 font-mono text-body text-dim"
	classSectionHead = "flex flex-wrap items-baseline justify-between gap-3 mt-10 mb-3"
	classEmpty       = "border border-dashed border-line rounded-[3px] p-10 text-center text-dim font-mono text-meta"
	classBadge       = "inline-block font-mono text-micro tracking-wider uppercase px-[0.5em] py-[0.15em] rounded-[3px] border border-current leading-normal"
	classNotice      = "font-ui text-meta px-4 py-3 rounded-[3px] border-l-2 border-danger bg-surface text-fg mb-4"
	classSkipLink    = "sr-only focus:not-sr-only focus:absolute focus:left-3 focus:top-2 focus:z-10 focus:bg-fg focus:text-page focus:px-3 focus:py-2 focus:rounded-[3px]"
)
