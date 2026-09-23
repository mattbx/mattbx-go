package ui

// Shared Tailwind utility clusters used across templates. These are an interim
// stand-in for @layer components / @apply once typography is locked — keep
// them as utility strings, not semantic CSS class names in main.css.

const (
	classPageHead = "mb-16"
	classEyebrow  = "font-mono text-[0.8125rem] tracking-[0.1em] uppercase text-muted m-0 mb-2"
	// classPageTitle matches the former --step-4 size used on detail pages.
	classPageTitle = "font-ui text-[2.75rem] font-semibold tracking-tight leading-[1.15] m-0 text-ink"
	// classPageTitleSm is the smaller index/section heading (Posts, Recent…).
	classPageTitleSm = "font-ui text-lg font-semibold tracking-tight m-0 text-ink"
	classPageLede    = "mt-2"
	classSectionHead = "flex flex-wrap items-baseline justify-between gap-3 mt-16 mb-3"
	classEmpty       = "border border-dashed border-muted/25 rounded-[3px] p-10 text-center text-muted font-ui text-sm"
	classBadge       = "inline-block font-mono text-[0.6875rem] tracking-wider uppercase px-[0.5em] py-[0.15em] rounded-[3px] border border-current leading-normal"
	classNotice      = "font-ui text-sm px-4 py-3 rounded-[3px] border-l-2 border-danger bg-surface text-ink mb-4"
	classSkipLink    = "sr-only focus:not-sr-only focus:absolute focus:left-3 focus:top-2 focus:z-10 focus:bg-ink focus:text-paper focus:px-3 focus:py-2 focus:rounded-[3px]"
)
