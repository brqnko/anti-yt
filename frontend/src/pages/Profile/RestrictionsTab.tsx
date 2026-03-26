import { useState, useEffect } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { getUser } from "../../api/generated/user";
import { getChannel } from "../../api/generated/channel";
import { getApiErrorCode } from "../../utils/api-error";
import { AddChannelDialog } from "../../components/AddChannelDialog";
import { TimeRangeSlider } from "../../components/TimeRangeSlider";
import { formatTime, parseTimeToMinutes } from "../../utils/format";
import type { TimeRange } from "../../types/time-range";
import type { GetChannelsSubscribed200ItemsItem } from "../../api/generated/antiYtApi.schemas";
import { Icon } from "../../components/Icon";

function RemoveChannelDialog({
  open,
  channel,
  onClose,
  onConfirm,
}: {
  open: boolean;
  channel: GetChannelsSubscribed200ItemsItem | null;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}) {
  const { t } = useTranslation();
  const [isRemoving, setIsRemoving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) {
      setIsRemoving(false);
      setError(null);
    }
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [open, onClose]);

  if (!open || !channel) return null;

  const handleConfirm = async () => {
    if (isRemoving) return;
    setIsRemoving(true);
    setError(null);
    try {
      await onConfirm();
      onClose();
    } catch {
      setError(t("restrictions.unsubscribeError"));
    } finally {
      setIsRemoving(false);
    }
  };

  return (
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/60" onClick={onClose} />
      <div class="relative bg-white dark:bg-[#2a2721] rounded-2xl ring-1 ring-black/10 dark:ring-white/10 border border-gray-100 dark:border-neutral-800 p-8 max-w-sm w-full">
        <button
          class="absolute top-4 right-4 text-text-muted-light dark:text-text-muted-dark hover:text-charcoal dark:hover:text-white transition-colors bg-transparent border-none cursor-pointer"
          onClick={onClose}
        >
          <Icon name="close" />
        </button>
        <div class="flex items-center gap-3 mb-4">
          <h2 class="text-lg font-bold text-charcoal dark:text-white">
            {t("restrictions.removeChannelDialog.title")}
          </h2>
        </div>
        <div class="flex items-center gap-3 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark mb-4">
          <img
            src={channel.external_channel_icon_url}
            alt=""
            loading="lazy"
            class="rounded-full size-10 shrink-0 border border-border-light dark:border-border-dark object-cover"
          />
          <div class="flex flex-col min-w-0">
            <p class="font-bold truncate text-sm">{channel.external_channel_display_name}</p>
            <p class="text-xs text-text-muted-light dark:text-text-muted-dark">{channel.channel_custom_id}</p>
          </div>
        </div>
        <p class="text-sm text-text-muted-light dark:text-text-muted-dark mb-4">
          {t("restrictions.removeChannelDialog.description", { name: channel.external_channel_display_name })}
        </p>
        {error && (
          <p class="text-sm text-red-500 mb-4">{error}</p>
        )}
        <div class="flex justify-end gap-3">
          <button
            class="px-4 py-2 rounded-xl text-sm font-medium text-text-muted-light dark:text-text-muted-dark hover:bg-black/5 dark:hover:bg-white/5 transition-colors bg-transparent border-none cursor-pointer"
            onClick={onClose}
          >
            {t("restrictions.removeChannelDialog.cancel")}
          </button>
          <button
            class="px-4 py-2 rounded-xl text-sm font-bold text-white bg-red-500 hover:bg-red-600 transition-colors border-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isRemoving}
            onClick={handleConfirm}
          >
            {isRemoving
              ? t("restrictions.removeChannelDialog.removing")
              : t("restrictions.removeChannelDialog.remove")}
          </button>
        </div>
      </div>
    </div>
  );
}

export function RestrictionsTab() {
  const { t } = useTranslation();

  // Time ranges from user's screen_time
  const [timeRanges, setTimeRanges] = useState<TimeRange[]>([]);

  // Daily cap
  const [hours, setHours] = useState(1);
  const [minutes, setMinutes] = useState(0);
  const [isUnlimited, setIsUnlimited] = useState(false);

  useEffect(() => {
    const { getUsersMeStatus } = getUser();
    getUsersMeStatus().then((user) => {
      setTimeRanges(
        (user.screen_time ?? []).map((slot) => ({
          id: slot.id,
          startMinutes: parseTimeToMinutes(slot.start_time),
          endMinutes: parseTimeToMinutes(slot.end_time),
        })),
      );
      const ds = user.daily_screen_seconds;
      if (ds != null && ds > 0) {
        setHours(Math.floor(ds / 3600));
        setMinutes(Math.floor((ds % 3600) / 60));
        setIsUnlimited(false);
      } else if (ds == null) {
        setIsUnlimited(true);
      }
    }).catch(() => {});
  }, []);

  // Saving state
  const [isSaving, setIsSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveFading, setSaveFading] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);

  // Whitelist (subscriptions)
  const [channels, setChannels] = useState<GetChannelsSubscribed200ItemsItem[]>([]);
  const [channelSearch, setChannelSearch] = useState("");
  const [showAddChannel, setShowAddChannel] = useState(false);

  useEffect(() => {
    const { getChannelsSubscribed } = getChannel();
    getChannelsSubscribed({ limit: 50 })
      .then((res) => setChannels(res.items))
      .catch(() => {});
  }, []);

  const filteredChannels = channels.filter((ch) =>
    ch.external_channel_display_name
      .toLowerCase()
      .includes(channelSearch.toLowerCase()),
  );

  const addRange = () => {
    setTimeRanges([
      ...timeRanges,
      { id: crypto.randomUUID(), startMinutes: 540, endMinutes: 660 },
    ]);
  };

  const updateRange = (id: string, updated: TimeRange) => {
    setTimeRanges(timeRanges.map((r) => (r.id === id ? updated : r)));
  };

  const deleteRange = (id: string) => {
    setTimeRanges(timeRanges.filter((r) => r.id !== id));
  };

  const handleSave = async () => {
    setIsSaving(true);
    setSaveSuccess(false);
    setSaveError(null);
    try {
      const { patchUsersMeStatus } = getUser();
      await patchUsersMeStatus({
        screen_time: timeRanges.map((r) => ({
          id: r.id,
          start_time: formatTime(r.startMinutes),
          end_time: formatTime(r.endMinutes),
        })),
        daily_screen_seconds: isUnlimited ? undefined : hours * 3600 + minutes * 60,
      });
      setSaveSuccess(true);
      setSaveFading(false);
      setTimeout(() => setSaveFading(true), 2500);
      setTimeout(() => { setSaveSuccess(false); setSaveFading(false); }, 3000);
    } catch (err) {
      const code = getApiErrorCode(err);
      setSaveError(code ? t(`apiErrors.${code}`, t("apiErrors.fallback")) : t("restrictions.saveError"));
    } finally {
      setIsSaving(false);
    }
  };

  const clampHours = (v: number) =>
    setHours(Math.max(0, Math.min(23, isNaN(v) ? 0 : v)));
  const clampMinutes = (v: number) =>
    setMinutes(Math.max(0, Math.min(59, isNaN(v) ? 0 : v)));

  const isTimeInvalid = !isUnlimited && hours === 0 && minutes === 0;

  const [removeTarget, setRemoveTarget] = useState<GetChannelsSubscribed200ItemsItem | null>(null);

  const handleUnsubscribe = async (channelId: string) => {
    const { deleteChannelsChannelIdSubscribe } = getChannel();
    await deleteChannelsChannelIdSubscribe(channelId);
    setChannels(channels.filter((ch) => ch.channel_id !== channelId));
  };

  return (
    <div class="flex flex-col gap-6 min-w-0 overflow-hidden">
      <div class="flex flex-col gap-2">
        <h1 class="text-3xl lg:text-4xl font-black leading-tight tracking-[-0.033em]">
          {t("restrictions.title")}
        </h1>
      </div>

      <div class="flex flex-col gap-6">
          {/* Time Constraints */}
          <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
            <div class="p-6 border-b border-border-light dark:border-border-dark">
              <h2 class="text-xl font-bold">
                {t("restrictions.timeConstraints")}
              </h2>
            </div>
            <div class="p-6 flex flex-col gap-8">
              {/* Permitted Hours */}
              <div class="flex flex-col gap-6">
                <div class="flex justify-between items-center flex-wrap gap-2">
                  <label class="text-base font-semibold">
                    {t("restrictions.permittedHours")}
                  </label>
                  <span class="text-xs font-medium text-text-muted-light dark:text-text-muted-dark">
                    {t("restrictions.permittedHoursDesc")}
                  </span>
                </div>
                <div class="flex flex-col gap-4">
                  {timeRanges.map((range) => (
                    <TimeRangeSlider
                      key={range.id}
                      range={range}
                      onChange={(updated) => updateRange(range.id, updated)}
                      onDelete={() => deleteRange(range.id)}
                    />
                  ))}
                </div>
                <button
                  type="button"
                  class="flex items-center justify-center gap-2 w-full py-3 border border-dashed border-border-light dark:border-border-dark rounded-lg text-text-muted-light dark:text-text-muted-dark hover:border-primary hover:text-primary hover:bg-primary/5 transition-all group cursor-pointer bg-transparent"
                  onClick={addRange}
                >
                  <Icon name="add" class="group-hover:scale-110 transition-transform" />
                  <span class="text-sm font-bold">
                    {t("restrictions.addTimeRange")}
                  </span>
                </button>
              </div>

              <hr class="border-border-light dark:border-border-dark" />

              {/* Daily Cap Limit */}
              <div class="flex flex-col gap-4">
                <label class="text-base font-semibold">
                  {t("restrictions.dailyCapLimit")}
                </label>
                <div class="flex flex-wrap gap-4 items-center">
                  {/* Unlimited toggle */}
                  <button
                    type="button"
                    role="switch"
                    aria-checked={!isUnlimited}
                    onClick={() => setIsUnlimited(!isUnlimited)}
                    class={`relative inline-flex h-7 w-12 shrink-0 items-center rounded-full transition-colors cursor-pointer border-none ${
                      !isUnlimited ? "bg-primary" : "bg-gray-300 dark:bg-gray-600"
                    }`}
                  >
                    <span
                      class={`inline-block size-5 rounded-full bg-white transition-transform ${
                        !isUnlimited ? "translate-x-6" : "translate-x-1"
                      }`}
                    />
                  </button>
                  <span class="text-sm font-medium text-text-muted-light dark:text-text-muted-dark">
                    {t("restrictions.enableLimit")}
                  </span>

                  <div class="w-px h-6 bg-border-light dark:bg-border-dark" />
                  <div class={`flex flex-wrap gap-4 items-center transition-opacity ${isUnlimited ? "opacity-40 pointer-events-none" : ""}`}>
                    <div class="relative">
                      <input
                        type="number"
                        class="w-24 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary outline-none text-center font-bold text-lg"
                        min={0}
                        max={23}
                        value={hours}
                        disabled={isUnlimited}
                        onInput={(e) =>
                          clampHours(
                            parseInt((e.target as HTMLInputElement).value),
                          )
                        }
                      />
                      <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-text-muted-light dark:text-text-muted-dark font-medium">
                        {t("restrictions.hr")}
                      </span>
                    </div>
                    <span class="text-text-muted-light dark:text-text-muted-dark font-bold">
                      :
                    </span>
                    <div class="relative">
                      <input
                        type="number"
                        class="w-24 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary outline-none text-center font-bold text-lg"
                        min={0}
                        max={59}
                        value={minutes}
                        disabled={isUnlimited}
                        onInput={(e) =>
                          clampMinutes(
                            parseInt((e.target as HTMLInputElement).value),
                          )
                        }
                      />
                      <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-text-muted-light dark:text-text-muted-dark font-medium">
                        {t("restrictions.min")}
                      </span>
                    </div>
                  </div>
                </div>
                {!isUnlimited && isTimeInvalid && (
                  <p class="text-sm text-red-500">
                    {t("restrictions.zeroTimeError")}
                  </p>
                )}
                <div class="flex items-center gap-3">
                  <button
                    class="px-6 py-3 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg transition-colors flex items-center gap-2 cursor-pointer border-none"
                    disabled={isSaving || isTimeInvalid}
                    onClick={handleSave}
                  >
                    {isSaving ? t("restrictions.saving") : t("restrictions.save")}
                  </button>
                  {saveSuccess && (
                    <span class={`text-sm text-green-600 dark:text-green-400 font-medium transition-opacity duration-500 ${saveFading ? "opacity-0" : "opacity-100"}`}>
                      {t("restrictions.saved")}
                    </span>
                  )}
                  {saveError && (
                    <span class="text-sm text-red-500 font-medium flex items-center gap-1">
                      <Icon name="error" class="text-[18px]" />
                      {saveError}
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>
          {/* Whitelist */}
          <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark border border-border-light dark:border-border-dark overflow-hidden">
            <div class="p-6 border-b border-border-light dark:border-border-dark">
              <h2 class="text-xl font-bold">
                {t("restrictions.whitelist")}
              </h2>
            </div>
            <div class="p-6 flex flex-col gap-6 grow">
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                  <Icon name="search" class="text-text-muted-light dark:text-text-muted-dark group-focus-within:text-primary transition-colors" />
                </div>
                <input
                  class="block w-full p-4 pl-10 text-sm bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary outline-none placeholder-text-muted-light dark:placeholder-text-muted-dark transition-all"
                  placeholder={t("restrictions.searchChannels")}
                  type="text"
                  value={channelSearch}
                  onInput={(e) =>
                    setChannelSearch((e.target as HTMLInputElement).value)
                  }
                />
              </div>
              <div class="flex flex-col gap-3 overflow-y-auto max-h-[400px] pr-1">
                {filteredChannels.length === 0 && (
                  <p class="text-sm text-text-muted-light dark:text-text-muted-dark text-center py-4">
                    {t("restrictions.noChannels")}
                  </p>
                )}
                {filteredChannels.map((ch) => (
                  <div
                    key={ch.channel_id}
                    class="flex items-center gap-4 p-3 rounded-lg bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark group hover:border-primary/50 transition-colors"
                  >
                    <img
                      src={ch.external_channel_icon_url}
                      alt=""
                      loading="lazy"
                      class="rounded-full size-12 shrink-0 border border-border-light dark:border-border-dark object-cover"
                    />
                    <div class="flex flex-col grow min-w-0">
                      <p class="font-bold truncate">
                        {ch.external_channel_display_name}
                      </p>
                      <p class="text-xs text-text-muted-light dark:text-text-muted-dark">
                        {ch.channel_custom_id}
                      </p>
                    </div>
                    <button
                      class="size-8 flex items-center justify-center rounded-full text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-all cursor-pointer bg-transparent border-none"
                      onClick={() => setRemoveTarget(ch)}
                    >
                      <Icon name="close" class="text-[20px]" />
                    </button>
                  </div>
                ))}
              </div>
              <button
                type="button"
                class="flex items-center justify-center gap-2 w-full py-3 border border-dashed border-border-light dark:border-border-dark rounded-lg text-text-muted-light dark:text-text-muted-dark hover:border-primary hover:text-primary hover:bg-primary/5 transition-all group cursor-pointer bg-transparent"
                onClick={() => setShowAddChannel(true)}
              >
                <Icon name="add" class="group-hover:scale-110 transition-transform" />
                <span class="text-sm font-bold">
                  {t("dashboard.requestChannel")}
                </span>
              </button>
            </div>
          </div>

      </div>

      <AddChannelDialog
        open={showAddChannel}
        onClose={() => setShowAddChannel(false)}
        onAdded={(sub) => setChannels((prev) => [...prev, sub])}
      />

      <RemoveChannelDialog
        open={removeTarget !== null}
        channel={removeTarget}
        onClose={() => setRemoveTarget(null)}
        onConfirm={() => handleUnsubscribe(removeTarget!.channel_id)}
      />
    </div>
  );
}
