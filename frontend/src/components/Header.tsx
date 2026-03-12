import { useLocation } from "preact-iso";
import { useTranslation } from "react-i18next";
import { useColorMode, type ColorMode } from "../hooks/useColorMode";

const modeIcons: Record<ColorMode, string> = {
  light: "\u2600\uFE0F",
  dark: "\uD83C\uDF19",
  system: "\uD83D\uDCBB",
};

const modeOrder: ColorMode[] = ["light", "dark", "system"];

export function Header() {
  const { url } = useLocation();
  const { t } = useTranslation();
  const { mode, setMode } = useColorMode();

  const cycleMode = () => {
    const next = modeOrder[(modeOrder.indexOf(mode) + 1) % modeOrder.length];
    setMode(next);
  };

  return (
    <header>
      <nav>
        <a href="/" class={url == "/" && "active"}>
          {t("common.home")}
        </a>
        <a href="/404" class={url == "/404" && "active"}>
          404
        </a>
      </nav>
      <button
        class="color-mode-btn"
        onClick={cycleMode}
        title={t(`common.colorMode.${mode}`)}
        aria-label={t(`common.colorMode.${mode}`)}
      >
        {modeIcons[mode]}
      </button>
    </header>
  );
}
