import { useState } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useTitle } from "../../hooks/useTitle";
import { ProtectedRoute } from "../../components/ProtectedRoute";
import { DashboardHeader } from "../../components/DashboardHeader";
import { useAuth } from "../../contexts/AuthContext";
import { getUser } from "../../api/generated/user";
import { languages } from "../../constants";
import { RestrictionsTab } from "./RestrictionsTab";
import { SecurityTab } from "./SecurityTab";

type Tab = "restrictions" | "profile" | "security";

function ProfileContent() {
  const { t, i18n } = useTranslation();
  const { user, logout, refreshUser } = useAuth();
  useTitle(t("profile.pageTitle"));

  const [activeTab, setActiveTab] = useState<Tab>("profile");

  const [displayName, setDisplayName] = useState(user?.display_name ?? "");
  const [languageCode, setLanguageCode] = useState(
    user?.language_code ?? (i18n.language.startsWith("ja") ? "ja" : "en"),
  );
  const [isSaving, setIsSaving] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [showLogoutConfirm, setShowLogoutConfirm] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);

  const handleSave = async () => {
    if (!displayName.trim() || displayName.trim().length < 3) return;
    setIsSaving(true);
    setSaveSuccess(false);
    try {
      const { patchUsersMeStatus } = getUser();
      await patchUsersMeStatus({
        display_name: displayName.trim(),
        language_code: languageCode,
      });
      await refreshUser();
      i18n.changeLanguage(languageCode);
      localStorage.setItem("lang", languageCode);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch {
      // Error handling could be improved
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
    <div class="relative flex min-h-screen w-full flex-col overflow-x-hidden bg-background-light dark:bg-background-dark text-charcoal dark:text-white font-display antialiased">
      <DashboardHeader />

      <div class="flex flex-1 w-full">
        {/* Sidebar */}
        <aside class="w-64 border-r border-border-light dark:border-border-dark hidden lg:flex flex-col p-6 gap-2">
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
          <div class="mt-auto pt-6 border-t border-border-light dark:border-border-dark">
            <button
              onClick={() => setShowLogoutConfirm(true)}
              class="flex items-center gap-3 px-4 py-3 w-full text-left rounded-xl hover:bg-red-50 dark:hover:bg-red-900/10 text-text-muted-light dark:text-text-muted-dark hover:text-red-500 transition-all font-bold cursor-pointer bg-transparent border-none"
            >
              <span class="material-symbols-outlined">logout</span>
              {t("profile.nav.signOut")}
            </button>
          </div>
        </aside>

        {/* Main Content */}
        <main class="flex-1 flex flex-col max-w-6xl w-full px-4 sm:px-6 lg:px-10 py-8 gap-8">
          {/* Mobile tab navigation */}
          <nav class="flex lg:hidden items-center gap-1 bg-background-light dark:bg-background-dark p-1 rounded-full border border-border-light dark:border-border-dark self-start">
            <button
              onClick={() => setActiveTab("profile")}
              class={`px-4 py-1.5 rounded-full text-sm font-bold transition-all cursor-pointer border-none ${
                activeTab === "profile"
                  ? "bg-card-light dark:bg-card-dark shadow-sm text-primary"
                  : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
              }`}
            >
              {t("profile.nav.profileSettings")}
            </button>
            <button
              onClick={() => setActiveTab("restrictions")}
              class={`px-4 py-1.5 rounded-full text-sm font-bold transition-all cursor-pointer border-none ${
                activeTab === "restrictions"
                  ? "bg-card-light dark:bg-card-dark shadow-sm text-primary"
                  : "bg-transparent text-text-muted-light dark:text-text-muted-dark hover:text-primary"
              }`}
            >
              {t("profile.nav.restrictions")}
            </button>
            <button
              onClick={() => setActiveTab("security")}
              class={`px-4 py-1.5 rounded-full text-sm font-bold transition-all cursor-pointer border-none ${
                activeTab === "security"
                  ? "bg-card-light dark:bg-card-dark shadow-sm text-primary"
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
              <div class="flex flex-col gap-2 mb-2">
                <h1 class="text-3xl lg:text-4xl font-black leading-tight tracking-[-0.033em]">
                  {t("profile.title")}
                </h1>
              </div>

              {/* Account Details */}
              <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
                <div class="p-6 border-b border-border-light dark:border-border-dark">
                  <h2 class="text-xl font-bold flex items-center gap-2">
                    <span class="material-symbols-outlined text-primary">
                      person_outline
                    </span>
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
                    <span class="text-sm text-green-600 dark:text-green-400 font-medium flex items-center gap-1">
                      <span class="material-symbols-outlined text-[18px]">
                        check_circle
                      </span>
                      {t("profile.saved")}
                    </span>
                  )}
                  <button
                    onClick={handleSave}
                    disabled={
                      isSaving ||
                      !displayName.trim() ||
                      displayName.trim().length < 3
                    }
                    class="px-8 py-2.5 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg transition-colors shadow-lg shadow-primary/20 cursor-pointer border-none"
                  >
                    {isSaving ? t("profile.saving") : t("profile.saveChanges")}
                  </button>
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
                    class="shrink-0 px-6 py-3 bg-red-600 hover:bg-red-700 text-white font-bold rounded-lg transition-all shadow-lg shadow-red-600/20 flex items-center justify-center gap-2 cursor-pointer border-none"
                  >
                    <span class="material-symbols-outlined text-[20px]">
                      delete_forever
                    </span>
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
        <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div class="bg-card-light dark:bg-card-dark rounded-xl shadow-2xl border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4">
            <div class="flex items-center gap-2 text-text-muted-light dark:text-text-muted-dark">
              <span class="material-symbols-outlined text-[28px]">
                logout
              </span>
              <h3 class="text-lg font-bold text-charcoal dark:text-white">
                {t("profile.logoutConfirmTitle")}
              </h3>
            </div>
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
                class="px-5 py-2.5 bg-red-600 hover:bg-red-700 text-white font-bold rounded-lg transition-colors cursor-pointer border-none flex items-center gap-2"
              >
                <span class="material-symbols-outlined text-[20px]">
                  logout
                </span>
                {t("profile.nav.signOut")}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {showDeleteConfirm && (
        <div class="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
          <div class="bg-card-light dark:bg-card-dark rounded-xl shadow-2xl border border-border-light dark:border-border-dark max-w-md w-full mx-4 p-6 flex flex-col gap-4">
            <div class="flex items-center gap-2 text-red-600 dark:text-red-400">
              <span class="material-symbols-outlined text-[28px]">
                warning
              </span>
              <h3 class="text-lg font-bold">
                {t("profile.deleteConfirmTitle")}
              </h3>
            </div>
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
