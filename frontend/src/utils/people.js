export function directoryPerson(directory, id, fallback = 'Unknown') {
  if (!id) return null
  const known = directory?.get?.(id)
  if (known) return known
  return { id, name: fallback }
}
