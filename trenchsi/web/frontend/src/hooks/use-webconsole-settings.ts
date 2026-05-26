import { useCallback, useEffect, useState } from "react"

export const WEBCONSOLE_STORAGE_KEY = "trenchsi.webconsole.settings"

export const WEBCONSOLE_FONT_SIZE_MIN = 10
export const WEBCONSOLE_FONT_SIZE_MAX = 22

export const WEBCONSOLE_COLOR_SCHEMES = {
  zinc: {
    labelKey: "pages.config.webconsole_color_zinc",
    background: "#09090b",
    border: "#27272a",
    empty: "#71717a",
    foreground: "#f4f4f5",
  },
  green: {
    labelKey: "pages.config.webconsole_color_green",
    background: "#03140c",
    border: "#166534",
    empty: "#4ade80",
    foreground: "#dcfce7",
  },
  amber: {
    labelKey: "pages.config.webconsole_color_amber",
    background: "#1c1204",
    border: "#92400e",
    empty: "#f59e0b",
    foreground: "#fffbeb",
  },
  slate: {
    labelKey: "pages.config.webconsole_color_slate",
    background: "#0f172a",
    border: "#334155",
    empty: "#94a3b8",
    foreground: "#f8fafc",
  },
} as const

export type WebconsoleColorScheme = keyof typeof WEBCONSOLE_COLOR_SCHEMES

export type WebconsoleSettings = {
  colorScheme: WebconsoleColorScheme
  fontSize: number
}

export const DEFAULT_WEBCONSOLE_SETTINGS: WebconsoleSettings = {
  colorScheme: "zinc",
  fontSize: 14,
}

export function normalizeWebconsoleSettings(
  settings: Partial<WebconsoleSettings> | null | undefined,
): WebconsoleSettings {
  const colorScheme =
    settings?.colorScheme && settings.colorScheme in WEBCONSOLE_COLOR_SCHEMES
      ? settings.colorScheme
      : DEFAULT_WEBCONSOLE_SETTINGS.colorScheme
  const numericFontSize = Number(settings?.fontSize)
  const fontSize = Number.isFinite(numericFontSize)
    ? Math.min(
        Math.max(Math.round(numericFontSize), WEBCONSOLE_FONT_SIZE_MIN),
        WEBCONSOLE_FONT_SIZE_MAX,
      )
    : DEFAULT_WEBCONSOLE_SETTINGS.fontSize

  return { colorScheme, fontSize }
}

export function readWebconsoleSettings(): WebconsoleSettings {
  if (typeof window === "undefined") {
    return DEFAULT_WEBCONSOLE_SETTINGS
  }

  try {
    const stored = window.localStorage.getItem(WEBCONSOLE_STORAGE_KEY)
    if (!stored) {
      return DEFAULT_WEBCONSOLE_SETTINGS
    }

    return normalizeWebconsoleSettings(JSON.parse(stored))
  } catch {
    return DEFAULT_WEBCONSOLE_SETTINGS
  }
}

export function writeWebconsoleSettings(settings: WebconsoleSettings) {
  if (typeof window === "undefined") {
    return
  }

  window.localStorage.setItem(
    WEBCONSOLE_STORAGE_KEY,
    JSON.stringify(normalizeWebconsoleSettings(settings)),
  )
}

export function useWebconsoleSettings() {
  const [settings, setSettings] = useState<WebconsoleSettings>(
    readWebconsoleSettings,
  )

  useEffect(() => {
    const handleStorage = (event: StorageEvent) => {
      if (event.key === WEBCONSOLE_STORAGE_KEY) {
        setSettings(readWebconsoleSettings())
      }
    }

    window.addEventListener("storage", handleStorage)
    return () => window.removeEventListener("storage", handleStorage)
  }, [])

  const updateSettings = useCallback((next: WebconsoleSettings) => {
    const normalized = normalizeWebconsoleSettings(next)
    writeWebconsoleSettings(normalized)
    setSettings(normalized)
  }, [])

  return { settings, updateSettings }
}
