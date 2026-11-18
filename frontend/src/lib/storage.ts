const KEY = "my-link-codes";
const MAX_ITEMS = 50;

export function loadCodes(): string[] {
  try {
    const parsed: unknown = JSON.parse(localStorage.getItem(KEY) ?? "[]");
    return Array.isArray(parsed)
      ? parsed.filter((c): c is string => typeof c === "string")
      : [];
  } catch {
    return [];
  }
}

function save(codes: string[]) {
  try {
    localStorage.setItem(KEY, JSON.stringify(codes.slice(0, MAX_ITEMS)));
  } catch {}
}

export function addCode(code: string) {
  save([code, ...loadCodes().filter((c) => c !== code)]);
}

export function removeCode(code: string) {
  save(loadCodes().filter((c) => c !== code));
}
