import { ref, watch } from "vue";

export type Theme = "system" | "light" | "dark";

const STORAGE_KEY = "garage-ui.theme";

function stored(): Theme {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value === "light" || value === "dark" ? value : "system";
  } catch {
    return "system";
  }
}

export const theme = ref<Theme>(stored());

watch(
  theme,
  (value) => {
    if (value === "system") document.documentElement.removeAttribute("data-theme");
    else document.documentElement.setAttribute("data-theme", value);
    try {
      localStorage.setItem(STORAGE_KEY, value);
    } catch {
      /* storage unavailable (private mode); theme stays per-session */
    }
  },
  { immediate: true },
);

export function cycleTheme() {
  theme.value = theme.value === "system" ? "light" : theme.value === "light" ? "dark" : "system";
}
