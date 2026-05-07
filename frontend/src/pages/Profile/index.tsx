import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardLayout } from "../../components/DashboardLayout";
import { useAuth } from "../../contexts/AuthContext";
import { getUser } from "../../api/generated/user";
import { getApiErrorCode } from "../../utils/api-error";
import { languages, modeIcons, REPORT_FORM_URL } from "../../constants";
import { useColorMode, type ColorMode } from "../../hooks/useColorMode";
import { RestrictionsTab } from "./RestrictionsTab";
import { SecurityTab } from "./SecurityTab";
import { HistoryTab } from "./HistoryTab";
import { Icon } from "../../components/Icon";
import { Dropdown } from "../../components/Dropdown";

type Tab = "profile" | "security" | "history";

function ProfileContent() {
  const { t, i18n } = useTranslation();
  const { logout } = useAuth();
  const { mode, setMode } = useColorMode();

  const colorModes: { value: ColorMode; icon: string }[] = [
    { value: "light", icon: modeIcons.light },
    { value: "dark", icon: modeIcons.dark },
    { value: "system", icon: modeIcons.system },
  ];
  useTitle(t("profile.pageTitle"));

  const [activeTab, setActiveTab] = useState<Tab>("profile");

  const [displayName, setDisplayName] = useState("");
  const [languageCode, setLanguageCode] = useState(
    languages.find((l) => l.code === i18n.language)?.code ?? "en",
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
  const [deleteConfirmInput, setDeleteConfirmInput] = useState("");
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveFading, setSaveFading] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [showImportDialog, setShowImportDialog] = useState(false);
  const [importSubscriptions, setImportSubscriptions] = useState(false);
  const [importLikes, setImportLikes] = useState(false);

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
    <DashboardLayout>
      <div class="flex-1 max-w-6xl w-full mx-auto px-4 sm:px-6 lg:px-10 py-8 flex flex-col gap-6">
        {/* Tab navigation */}
        <nav class="flex items-center gap-1 bg-background-light dark:bg-background-dark p-1 rounded-full border border-border-light dark:border-border-dark overflow-x-auto">
          <button
            onClick={() => setActiveTab("profile")}
            class={`flex-1 px-4 py-1.5 rounded-full text-sm font-bold cursor-pointer border-none whitespace-nowrap ${
              activeTab === "profile"
                ? "bg-card-light dark:bg-card-dark text-primary"
                : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
            }`}
          >
            {t("profile.nav.profileSettings")}
          </button>
          <button
            onClick={() => setActiveTab("security")}
            class={`flex-1 px-4 py-1.5 rounded-full text-sm font-bold cursor-pointer border-none whitespace-nowrap ${
              activeTab === "security"
                ? "bg-card-light dark:bg-card-dark text-primary"
                : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
            }`}
          >
            {t("profile.nav.security")}
          </button>
          <button
            onClick={() => setActiveTab("history")}
            class={`flex-1 px-4 py-1.5 rounded-full text-sm font-bold cursor-pointer border-none whitespace-nowrap ${
              activeTab === "history"
                ? "bg-card-light dark:bg-card-dark text-primary"
                : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
            }`}
          >
            {t("dashboard.nav.history")}
          </button>
        </nav>

        {activeTab === "security" && <SecurityTab />}

        {activeTab === "history" && <HistoryTab />}

        {activeTab === "profile" && (
          <>
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
                      <Icon name="edit" class="text-[20px]" />
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
                  <div class="mt-1 flex items-center justify-end">
                    <span class={`text-xs tabular-nums ${nameLength > 29 ? "text-red-500" : "text-text-muted-light dark:text-text-muted-dark"}`}>
                      {nameLength}/29
                    </span>
                  </div>
                </div>
                <div class="flex flex-col gap-2">
                  <label class="text-sm font-bold text-text-muted-light dark:text-text-muted-dark uppercase tracking-wider">
                    {t("profile.language")}
                  </label>
                  <Dropdown
                    value={languageCode}
                    onChange={setLanguageCode}
                    ariaLabel={t("profile.language")}
                    leadingIcon="language"
                    options={languages.map((lang) => ({
                      value: lang.code,
                      label: lang.label,
                    }))}
                  />
                </div>
              </div>
              <div class="px-6 py-4 border-t border-border-light dark:border-border-dark flex items-center">
                <button
                  onClick={() => setShowImportDialog(true)}
                  class="flex items-center justify-center gap-2.5 rounded-xl bg-white dark:bg-[#242424] px-5 py-2.5 text-sm font-bold text-slate-700 dark:text-white border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-[#2a2a2a] hover:border-[#FF0000]/50 focus:outline-none focus:ring-2 focus:ring-[#FF0000] focus:ring-offset-2 dark:focus:ring-offset-[var(--color-bg)] cursor-pointer"
                >
                  <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814z" fill="#FF0000"/>
                    <path d="M9.545 15.568V8.432L15.818 12l-6.273 3.568z" fill="#ffffff"/>
                  </svg>
                  {t("profile.youtubeImport.button")}
                </button>
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
                  class="px-8 py-2.5 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg cursor-pointer border-none"
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
                  <Dropdown
                    value={mode}
                    onChange={(v) => setMode(v as ColorMode)}
                    ariaLabel={t("appearance.colorMode")}
                    leadingIcon={modeIcons[mode]}
                    options={colorModes.map((m) => ({
                      value: m.value,
                      label: t(`common.colorMode.${m.value}`),
                    }))}
                  />
                </div>
              </div>
            </div>

            {/* Restrictions */}
            <RestrictionsTab />

            {/* その他 */}
            <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
              <div class="px-6 py-4 border-b border-border-light dark:border-border-dark">
                <h2 class="text-xl font-bold">
                  {t("profile.nav.other")}
                </h2>
              </div>
              <div class="p-6 flex flex-col gap-1">
                <a
                  href={REPORT_FORM_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  class="flex items-center gap-3 px-4 py-3 w-full text-left rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white font-bold no-underline"
                >
                  <Icon name="flag" />
                  {t("profile.nav.reportProblem")}
                </a>
                <div class="mt-4" />
                <button
                  onClick={() => setShowLogoutConfirm(true)}
                  class="flex items-center gap-3 px-4 py-3 w-full text-left rounded-xl hover:bg-black/5 dark:hover:bg-white/5 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white font-bold cursor-pointer bg-transparent border-none"
                >
                  <Icon name="logout" />
                  {t("profile.nav.signOut")}
                </button>
                <div class="mt-4" />
                <button
                  onClick={() => setShowDeleteConfirm(true)}
                  class="flex items-center gap-3 px-4 py-3 w-full text-left rounded-xl hover:bg-red-50 dark:hover:bg-red-900/10 text-red-500 font-bold cursor-pointer bg-transparent border-none"
                >
                  <Icon name="delete_forever" />
                  {t("profile.deleteAccount")}
                </button>
              </div>
            </div>

          </>
        )}
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
                class="px-5 py-2.5 rounded-lg font-bold text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
              >
                {t("profile.cancel")}
              </button>
              <button
                onClick={() => logout()}
                class="px-5 py-2.5 bg-red-600 hover:bg-red-700 text-white font-bold rounded-lg cursor-pointer border-none"
              >
                {t("profile.nav.signOut")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteConfirm && (
        <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60" onClick={() => { if (!isDeleting) { setShowDeleteConfirm(false); setDeleteConfirmInput(""); } }}>
          <div class="bg-card-light dark:bg-card-dark rounded-xl ring-1 ring-black/10 dark:ring-white/10 border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4" onClick={(e) => e.stopPropagation()}>
            <h3 class="text-lg font-bold text-red-600 dark:text-red-400">
              {t("profile.deleteConfirmTitle")}
            </h3>
            <p class="text-sm text-text-muted-light dark:text-text-muted-dark leading-relaxed">
              {t("profile.deleteConfirmDesc")}
            </p>
            <div class="flex flex-col gap-2">
              <label class="text-sm font-medium text-text-muted-light dark:text-text-muted-dark">
                {t("profile.deleteConfirmInputLabel")}
              </label>
              <input
                type="text"
                value={deleteConfirmInput}
                onInput={(e) => setDeleteConfirmInput((e.target as HTMLInputElement).value)}
                placeholder="Delete my account"
                class="w-full px-4 py-2.5 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-red-500 focus:border-transparent outline-none transition-all font-medium"
                autocomplete="off"
              />
            </div>
            <div class="flex justify-end gap-3 mt-2">
              <button
                onClick={() => { setShowDeleteConfirm(false); setDeleteConfirmInput(""); }}
                disabled={isDeleting}
                class="px-5 py-2.5 rounded-lg font-bold text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
              >
                {t("profile.cancel")}
              </button>
              <button
                onClick={handleDelete}
                disabled={isDeleting || deleteConfirmInput !== "Delete my account"}
                class="px-5 py-2.5 bg-red-600 hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg cursor-pointer border-none flex items-center gap-2"
              >
                {isDeleting && (
                  <Icon name="progress_activity" class="text-[18px] animate-spin" />
                )}
                {t("profile.deleteConfirm")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* YouTube Import Dialog */}
      {showImportDialog && (
        <div
          class="fixed inset-0 z-[100] flex items-center justify-center bg-black/60"
          onClick={() => setShowImportDialog(false)}
        >
          <div
            class="bg-card-light dark:bg-card-dark rounded-xl ring-1 ring-black/10 dark:ring-white/10 border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-5"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 class="text-lg font-bold text-charcoal dark:text-white">
              {t("profile.youtubeImport.dialogTitle")}
            </h3>

            <div class="flex flex-col gap-3">
              <label class="flex items-start gap-3 p-3 rounded-lg border border-border-light dark:border-border-dark hover:border-primary/30 cursor-pointer">
                <input
                  type="checkbox"
                  checked={importSubscriptions}
                  onChange={() => setImportSubscriptions((v) => !v)}
                  class="mt-0.5 size-4 accent-primary cursor-pointer"
                />
                <div class="flex flex-col">
                  <span class="text-sm font-bold text-charcoal dark:text-white">
                    {t("profile.youtubeImport.subscriptions")}
                  </span>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark mt-0.5">
                    {t("profile.youtubeImport.subscriptionsDesc")}
                  </span>
                </div>
              </label>

              <label class="flex items-start gap-3 p-3 rounded-lg border border-border-light dark:border-border-dark hover:border-primary/30 cursor-pointer">
                <input
                  type="checkbox"
                  checked={importLikes}
                  onChange={() => setImportLikes((v) => !v)}
                  class="mt-0.5 size-4 accent-primary cursor-pointer"
                />
                <div class="flex flex-col">
                  <span class="text-sm font-bold text-charcoal dark:text-white">
                    {t("profile.youtubeImport.likes")}
                  </span>
                  <span class="text-xs text-text-muted-light dark:text-text-muted-dark mt-0.5">
                    {t("profile.youtubeImport.likesDesc")}
                  </span>
                </div>
              </label>
            </div>

            {!importSubscriptions && !importLikes && (
              <p class="text-xs text-red-500 font-medium">
                {t("profile.youtubeImport.selectAtLeastOne")}
              </p>
            )}

            <div class="flex justify-end gap-3">
              <button
                onClick={() => setShowImportDialog(false)}
                class="px-5 py-2.5 rounded-lg font-bold text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer bg-transparent border-none"
              >
                {t("profile.youtubeImport.cancel")}
              </button>
              <button
                disabled={!importSubscriptions && !importLikes}
                onClick={() => {
                  const params = new URLSearchParams();
                  if (importSubscriptions) params.set("subscriptions", "true");
                  if (importLikes) params.set("likes", "true");
                  window.location.href = `/api/v1/auth/oauth/youtube?${params.toString()}`;
                }}
                class="px-5 py-2.5 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg cursor-pointer border-none"
              >
                {t("profile.youtubeImport.start")}
              </button>
            </div>
          </div>
        </div>
      )}
    </DashboardLayout>
  );
}

export default function Profile() {
  return (
    <ProtectedRoute>
      <ProfileContent />
    </ProtectedRoute>
  );
}
