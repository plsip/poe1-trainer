export function formatGuideVersion(version: string): string {
  const trimmed = version.trim()
  if (/^[0-9a-f]{40}$/i.test(trimmed)) {
    return trimmed.slice(0, 7)
  }
  return trimmed
}