const GUIDE_SLUG_ALIASES: Record<string, string> = {
  stormburst_campaign_v1: 'stormburst_campaign',
}

export function getCanonicalGuideSlug(slug?: string): string | undefined {
  if (!slug) return slug
  return GUIDE_SLUG_ALIASES[slug] ?? slug
}
