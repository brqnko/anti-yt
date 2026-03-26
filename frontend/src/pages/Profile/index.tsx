import { useState, useEffect, useCallback } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardHeader } from "../../components/DashboardHeader";
import { useAuth } from "../../contexts/AuthContext";
import { getUser } from "../../api/generated/user";
import { getApiErrorCode } from "../../utils/api-error";
import { languages, modeIcons, REPORT_FORM_URL } from "../../constants";
import { useColorMode, type ColorMode } from "../../hooks/useColorMode";
import { RestrictionsTab } from "./RestrictionsTab";
import { SecurityTab } from "./SecurityTab";

const SIDEBAR_STORAGE_KEY = "sidebar-open";

function getStoredSidebarState(): boolean {
  try {
    const stored = localStorage.getItem(SIDEBAR_STORAGE_KEY);
    if (stored !== null) return stored === "true";
  } catch {}
  return true;
}

type Tab = "restrictions" | "profile" | "security";

function ProfileContent() {
  const { t, i18n } = useTranslation();
  const { logout } = useAuth();
  const { mode, setMode } = useColorMode();

  const [sidebarOpen, setSidebarOpen] = useState(getStoredSidebarState);

  const toggleSidebar = useCallback(() => {
    setSidebarOpen((v) => {
      const next = !v;
      try { localStorage.setItem(SIDEBAR_STORAGE_KEY, String(next)); } catch {}
      return next;
    });
  }, []);

  const colorModes: { value: ColorMode; icon: string }[] = [
    { value: "light", icon: modeIcons.light },
    { value: "dark", icon: modeIcons.dark },
    { value: "system", icon: modeIcons.system },
  ];
  useTitle(t("profile.pageTitle"));

  const [activeTab, setActiveTab] = useState<Tab>("profile");

  const [displayName, setDisplayName] = useState("");
  const [languageCode, setLanguageCode] = useState(
    i18n.language.startsWith("ja") ? "ja" : "en",
  );

  useEffect(() => {
    const { getUsersMeStatus } = getUser();
    getUsersMeStatus().then((user) => {
      setDisplayName(user.display_name ?? "");
      setLanguageCode(user.language_code);
    }).catch(() => {});
  }, []);
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveFading, setSaveFading] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  const trimmedName = displayName.trim();
  const nameLength = [...trimmedName].length;
  const isNameValid = nameLength >= 3 && nameLength <= 29;

  const handleSave = async () => {
    if (!isNameValid) return;
    setIsSaving(true);
    setSaveSuccess(false);
    setSaveError(null);
    try {
      const { patchUsersMeStatus } = getUser();
      await patchUsersMeStatus({
        display_name: displayName.trim(),
        language_code: languageCode,
      });
      i18n.changeLanguage(languageCode);
      localStorage.setItem("lang", languageCode);
      setSaveSuccess(true);
      setSaveFading(false);
      setTimeout(() => setSaveFading(true), 2500);
      setTimeout(() => { setSaveSuccess(false); setSaveFading(false); }, 3000);
    } catch (err) {
      const code = getApiErrorCode(err);
      setSaveError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("apiErrors.fallback"));
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async () => {
    setIsDeleting(true);
    try {
      const { deleteUsersMe } = getUser();
      await deleteUsersMe();
      await logout();
    } catch {
      setIsDeleting(false);
      setShowDeleteConfirm(false);
    }
  };

  return (
    <div class="relative flex h-screen w-full flex-col overflow-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader sidebarOpen={sidebarOpen} onToggleSidebar={toggleSidebar} />

      <div class="flex flex-1 w-full max-w-[1600px] mx-auto overflow-hidden">
        {/* Mobile backdrop */}
        {sidebarOpen && (
          <div
            class="fixed inset-0 z-30 bg-black/40 lg:hidden"
            onClick={toggleSidebar}
          />
        )}

        {/* Sidebar */}
        <aside
          class={`flex flex-col border-r border-border-light dark:border-border-dark shrink-0 transition-[width,opacity,transform] duration-200
            fixed top-[57px] bottom-0 z-40 bg-background-light dark:bg-background-dark
            lg:relative lg:top-auto lg:bottom-auto lg:z-auto
            ${sidebarOpen
              ? "w-64 opacity-100 translate-x-0 overflow-y-auto overflow-x-hidden"
              : "w-0 opacity-0 -translate-x-full lg:translate-x-0 overflow-hidden"
            }`}
        >
          <div class="flex flex-col flex-1 p-6 gap-2 min-w-[16rem]">
          <button
            onClick={() => setActiveTab("profile")}
            class={`flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer transition-all font-bold border-none bg-transparent w-full text-left ${
              activeTab === "profile"
                ? "bg-primary/10 text-primary"
                : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5"
            }`}
          >
            <span class="material-symbols-outlined">person</span>
            {t("profile.nav.profileSettings")}
          </button>
          <button
            onClick={() => setActiveTab("restrictions")}
            class={`flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer transition-all font-bold border-none bg-transparent w-full text-left ${
              activeTab === "restrictions"
                ? "bg-primary/10 text-primary"
                : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5"
            }`}
          >
            <span class="material-symbols-outlined">block</span>
            {t("profile.nav.restrictions")}
          </button>
          <button
            onClick={() => setActiveTab("security")}
            class={`flex items-center gap-3 px-4 py-3 rounded-xl cursor-pointer transition-all font-bold border-none bg-transparent w-full text-left ${
              activeTab === "security"
                ? "bg-primary/10 text-primary"
                : "text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5"
            }`}
          >
            <span class="material-symbols-outlined">shield_lock</span>
            {t("profile.nav.security")}
          </button>
          <div class="mt-auto flex flex-col gap-1">
            <a
              href={REPORT_FORM_URL}
              target="_blank"
              rel="noopener noreferrer"
              class="flex items-center gap-3 px-4 py-3 w-full text-left rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-all font-bold no-underline"
            >
              <span class="material-symbols-outlined">flag</span>
              {t("profile.nav.reportProblem")}
            </a>
            <div class="border-t border-border-light dark:border-border-dark my-1" />
            <button
              onClick={() => setShowLogoutConfirm(true)}
              class="flex items-center gap-3 px-4 py-3 w-full text-left rounded-xl hover:bg-red-50 dark:hover:bg-red-900/10 text-text-muted-light dark:text-text-muted-dark hover:text-red-500 transition-all font-bold cursor-pointer bg-transparent border-none"
            >
              <span class="material-symbols-outlined">logout</span>
              {t("profile.nav.signOut")}
            </button>
          </div>
          </div>
        </aside>

        {/* Main Content */}
        <main class="flex-1 flex flex-col max-w-6xl w-full px-4 sm:px-6 lg:px-10 py-8 gap-6 overflow-y-auto">
          {/* Mobile tab navigation */}
          <nav class="flex lg:hidden items-center gap-1 bg-background-light dark:bg-background-dark p-1 rounded-full border border-border-light dark:border-border-dark self-start overflow-x-auto">
            <button
              onClick={() => setActiveTab("profile")}
              class={`px-4 py-1.5 rounded-full text-sm font-bold transition-all cursor-pointer border-none whitespace-nowrap ${
                activeTab === "profile"
                  ? "bg-card-light dark:bg-card-dark text-primary"
                  : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
              }`}
            >
              {t("profile.nav.profileSettings")}
            </button>
            <button
              onClick={() => setActiveTab("restrictions")}
              class={`px-4 py-1.5 rounded-full text-sm font-bold transition-all cursor-pointer border-none whitespace-nowrap ${
                activeTab === "restrictions"
                  ? "bg-card-light dark:bg-card-dark text-primary"
                  : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
              }`}
            >
              {t("profile.nav.restrictions")}
            </button>
            <button
              onClick={() => setActiveTab("security")}
              class={`px-4 py-1.5 rounded-full text-sm font-bold transition-all cursor-pointer border-none whitespace-nowrap ${
                activeTab === "security"
                  ? "bg-card-light dark:bg-card-dark text-primary"
                  : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
              }`}
            >
              {t("profile.nav.security")}
            </button>
          </nav>

          {activeTab === "restrictions" && <RestrictionsTab />}

          {activeTab === "security" && <SecurityTab />}

          {activeTab === "profile" && (
            <>
              <div class="flex flex-col">
                <h1 class="text-3xl lg:text-4xl font-black leading-tight tracking-[-0.033em]">
                  {t("profile.title")}
                </h1>
              </div>

              {/* Account Details */}
              <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
                <div class="px-6 py-4 border-b border-border-light dark:border-border-dark">
                  <h2 class="text-xl font-bold">
                    {t("profile.accountDetails")}
                  </h2>
                </div>
                <div class="p-6 grid grid-cols-1 sm:grid-cols-2 gap-6">
                  <div class="flex flex-col gap-2">
                    <label class="text-sm font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-wider">
                      {t("profile.displayName")}
                    </label>
                    <div class="relative">
                      <span class="absolute inset-y-0 left-3 flex items-center text-text-muted-light dark:text-text-muted-dark">
                        <span class="material-symbols-outlined text-[20px]">
                          edit
                        </span>
                      </span>
                      <input
                        type="text"
                        value={displayName}
                        onInput={(e) =>
                          setDisplayName((e.target as HTMLInputElement).value)
                        }
                        placeholder={t("profile.displayNamePlaceholder")}
                        class="w-full pl-10 pr-4 py-2.5 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none transition-all font-medium"
                      />
                    </div>
                    <div class="mt-1 flex items-center justify-between">
                      <p class="text-xs text-text-muted-light dark:text-text-muted-dark">
                        {t("register.profileDetails.displayNameHint")}
                      </p>
                      <span class={`text-xs tabular-nums ${nameLength > 29 ? "text-red-500" : "text-text-muted-light dark:text-text-muted-dark"}`}>
                        {nameLength}/29
                      </span>
                    </div>
                  </div>
                  <div class="flex flex-col gap-2">
                    <label class="text-sm font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-wider">
                      {t("profile.language")}
                    </label>
                    <div class="relative">
                      <span class="absolute inset-y-0 left-3 flex items-center text-text-muted-light dark:text-text-muted-dark pointer-events-none">
                        <span class="material-symbols-outlined text-[20px]">
                          language
                        </span>
                      </span>
                      <select
                        value={languageCode}
                        onChange={(e) =>
                          setLanguageCode(
                            (e.target as HTMLSelectElement).value,
                          )
                        }
                        class="w-full pl-10 pr-8 py-2.5 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none appearance-none font-medium"
                      >
                        {languages.map((lang) => (
                          <option key={lang.code} value={lang.code}>
                            {lang.label}
                          </option>
                        ))}
                      </select>
                      <span class="absolute inset-y-0 right-3 flex items-center text-text-muted-light dark:text-text-muted-dark pointer-events-none">
                        <span class="material-symbols-outlined text-[20px]">
                          expand_more
                        </span>
                      </span>
                    </div>
                  </div>
                </div>
                <div class="px-6 py-4 bg-background-light/50 dark:bg-background-dark/50 border-t border-border-light dark:border-border-dark flex items-center justify-end gap-3">
                  {saveSuccess && (
                    <span class={`text-sm text-green-600 dark:text-green-400 font-medium transition-opacity duration-500 ${saveFading ? "opacity-0" : "opacity-100"}`}>
                      {t("profile.saved")}
                    </span>
                  )}
                  {saveError && (
                    <span class="text-sm text-red-500 font-medium">
                      {saveError}
                    </span>
                  )}
                  <button
                    onClick={handleSave}
                    disabled={isSaving || !isNameValid}
                    class="px-8 py-2.5 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg transition-colors cursor-pointer border-none"
                  >
                    {isSaving ? t("profile.saving") : t("profile.saveChanges")}
                  </button>
                </div>
              </div>

              {/* Appearance */}
              <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
                <div class="px-6 py-4 border-b border-border-light dark:border-border-dark">
                  <h2 class="text-xl font-bold">
                    {t("appearance.title")}
                  </h2>
                </div>
                <div class="p-6">
                  <div class="flex flex-col gap-2 max-w-xs">
                    <label class="text-sm font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-wider">
                      {t("appearance.colorMode")}
                    </label>
                    <div class="relative">
                      <span class="absolute inset-y-0 left-3 flex items-center text-text-muted-light dark:text-text-muted-dark pointer-events-none">
                        <span class="material-symbols-outlined text-[20px]">
                          {colorModes.find((m) => m.value === mode)?.icon}
                        </span>
                      </span>
                      <select
                        value={mode}
                        onChange={(e) => setMode((e.target as HTMLSelectElement).value as ColorMode)}
                        class="w-full pl-10 pr-8 py-2.5 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary focus:border-transparent outline-none appearance-none font-medium"
                      >
                        {colorModes.map((m) => (
                          <option key={m.value} value={m.value}>
                            {t(`common.colorMode.${m.value}`)}
                          </option>
                        ))}
                      </select>
                      <span class="absolute inset-y-0 right-3 flex items-center text-text-muted-light dark:text-text-muted-dark pointer-events-none">
                        <span class="material-symbols-outlined text-[20px]">
                          expand_more
                        </span>
                      </span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Danger Zone */}
              <div class="rounded-xl border border-red-200 dark:border-red-900/30 bg-red-50/30 dark:bg-red-950/20 overflow-hidden">
                <div class="p-6 flex flex-col sm:flex-row sm:items-center justify-between gap-6">
                  <div class="max-w-xl">
                    <h3 class="font-bold mb-1">
                      {t("profile.deleteAccount")}
                    </h3>
                    <p class="text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed">
                      {t("profile.deleteAccountDesc")}
                    </p>
                  </div>
                  <button
                    onClick={() => setShowDeleteConfirm(true)}
                    class="shrink-0 px-6 py-3 bg-red-600 hover:bg-red-700 text-white font-bold rounded-lg transition-all flex items-center justify-center cursor-pointer border-none"
                  >
                    {t("profile.deleteAccount")}
                  </button>
                </div>
              </div>
            </>
          )}
        </main>
      </div>

      {/* Logout Confirmation Dialog */}
      {showLogoutConfirm && (
        <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60" onClick={() => setShowLogoutConfirm(false)}>
          <div class="bg-card-light dark:bg-card-dark rounded-xl ring-1 ring-black/10 dark:ring-white/10 border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4" onClick={(e) => e.stopPropagation()}>
            <h3 class="text-lg font-bold text-charcoal dark:text-white">
              {t("profile.logoutConfirmTitle")}
            </h3>
            <p class="text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed">
              {t("profile.logoutConfirmDesc")}
            </p>
            <div class="flex justify-end gap-3 mt-2">
              <button
                onClick={() => setShowLogoutConfirm(false)}
                class="px-5 py-2.5 rounded-lg font-bold text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
              >
                {t("profile.cancel")}
              </button>
              <button
                onClick={() => logout()}
                class="px-5 py-2.5 bg-red-600 hover:bg-red-700 text-white font-bold rounded-lg transition-colors cursor-pointer border-none"
              >
                {t("profile.nav.signOut")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteConfirm && (
        <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60" onClick={() => { if (!isDeleting) setShowDeleteConfirm(false); }}>
          <div class="bg-card-light dark:bg-card-dark rounded-xl ring-1 ring-black/10 dark:ring-white/10 border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4" onClick={(e) => e.stopPropagation()}>
            <h3 class="text-lg font-bold text-red-600 dark:text-red-400">
              {t("profile.deleteConfirmTitle")}
            </h3>
            <p class="text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed">
              {t("profile.deleteConfirmDesc")}
            </p>
            <div class="flex justify-end gap-3 mt-2">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                disabled={isDeleting}
                class="px-5 py-2.5 rounded-lg font-bold text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer bg-transparent border-none"
              >
                {t("profile.cancel")}
              </button>
              <button
                onClick={handleDelete}
                disabled={isDeleting}
                class="px-5 py-2.5 bg-red-600 hover:bg-red-700 disabled:opacity-50 text-white font-bold rounded-lg transition-colors cursor-pointer border-none flex items-center gap-2"
              >
                {isDeleting && (
                  <span class="material-symbols-outlined text-[18px] animate-spin">
                    progress_activity
                  </span>
                )}
                {t("profile.deleteConfirm")}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default function Profile() {
  return (
    <ProtectedRoute>
      <ProfileContent />
    </ProtectedRoute>
  );
}
