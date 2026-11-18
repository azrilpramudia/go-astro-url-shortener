export async function copyToClipboard(text: string) {
  const clipboard = navigator.clipboard;
  if (!clipboard?.writeText) {
    throw new Error("Clipboard access is unavailable.");
  }

  await clipboard.writeText(text);
}
