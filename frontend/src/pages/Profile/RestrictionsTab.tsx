import { useState, useEffect, useRef } from "preact/hooks";
import { useTranslation } from "react-i18next";
import { useAuth } from "../../contexts/AuthContext";
import { getUser } from "../../api/generated/user";
import { getChannel } from "../../api/generated/channel";
import type { GetSubscriptions200ItemsItem } from "../../api/generated/antiYtApi.schemas";

interface TimeRange {
  id: string;
  startMinutes: number;
  endMinutes: number;
}

function formatTime(minutes: number): string {
  const h = Math.floor(minutes / 60)
    .toString()
    .padStart(2, "0");
  const m = (minutes % 60).toString().padStart(2, "0");
  return `${h}:${m}`;
}

function parseTimeToMinutes(time: string): number {
  const [h, m] = time.split(":").map(Number);
  return h * 60 + m;
}

function TimeRangeSlider({
  range,
  onChange,
  onDelete,
}: {
  range: TimeRange;
  onChange: (r: TimeRange) => void;
  onDelete: () => void;
}) {
  const { t } = useTranslation();
  const trackRef = useRef<HTMLDivElement>(null);
  const draggingRef = useRef<"start" | "end" | null>(null);

  const totalMinutes = 1440;
  const startPct = (range.startMinutes / totalMinutes) * 100;
  const endPct = (range.endMinutes / totalMinutes) * 100;

  const calcMinutes = (clientX: number): number => {
    if (!trackRef.current) return 0;
    const rect = trackRef.current.getBoundingClientRect();
    const x = Math.max(0, Math.min(clientX - rect.left, rect.width));
    const raw = (x / rect.width) * totalMinutes;
    return Math.max(0, Math.min(Math.round(raw / 15) * 15, totalMinutes));
  };

  const handlePointerDown =
    (thumb: "start" | "end") => (e: PointerEvent) => {
      e.preventDefault();
      (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
      draggingRef.current = thumb;
    };

  const handlePointerMove = (e: PointerEvent) => {
    const thumb = draggingRef.current;
    if (!thumb) return;
    const mins = calcMinutes(e.clientX);
    if (thumb === "start") {
      onChange({
        ...range,
        startMinutes: Math.min(mins, range.endMinutes - 15),
      });
    } else {
      onChange({
        ...range,
        endMinutes: Math.max(mins, range.startMinutes + 15),
      });
    }
  };

  const handlePointerUp = (e: PointerEvent) => {
    (e.currentTarget as HTMLElement).releasePointerCapture(e.pointerId);
    draggingRef.current = null;
  };

  const thumbClass =
    "absolute top-1/2 -mt-2 -ml-2 size-4 bg-white dark:bg-card-dark border-2 border-primary rounded-full shadow-md cursor-grab z-10 hover:scale-110 active:scale-95 transition-transform touch-none";

  return (
    <div class="flex flex-col gap-4 p-4 rounded-xl bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="px-2.5 py-1 bg-primary/10 text-primary text-xs font-bold rounded border border-primary/20">
            {formatTime(range.startMinutes)} - {formatTime(range.endMinutes)}
          </div>
        </div>
        <button
          type="button"
          aria-label={t("restrictions.removeRange")}
          class="size-8 flex items-center justify-center text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-all cursor-pointer bg-transparent border-none"
          onClick={onDelete}
        >
          <span class="material-symbols-outlined text-[18px]">delete</span>
        </button>
      </div>
      <div class="relative pt-2 px-1">
        <div class="flex justify-between text-[10px] text-text-muted-light dark:text-text-muted-dark mb-2 font-medium">
          <span>00:00</span>
          <span class="hidden sm:inline">06:00</span>
          <span>12:00</span>
          <span class="hidden sm:inline">18:00</span>
          <span>24:00</span>
        </div>
        <div
          ref={trackRef}
          class="h-1.5 w-full bg-border-light dark:bg-border-dark rounded-full relative"
        >
          <div
            class="absolute h-full bg-primary rounded-full pointer-events-none"
            style={{
              left: `${startPct}%`,
              width: `${endPct - startPct}%`,
            }}
          />
          <div
            class={thumbClass}
            style={{ left: `${startPct}%` }}
            onPointerDown={handlePointerDown("start")}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
          />
          <div
            class={thumbClass}
            style={{ left: `${endPct}%` }}
            onPointerDown={handlePointerDown("end")}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerUp}
          />
        </div>
      </div>
    </div>
  );
}

export function RestrictionsTab() {
  const { t } = useTranslation();
  const { user, refreshUser } = useAuth();

  // Time ranges from user's screen_time
  const [timeRanges, setTimeRanges] = useState<TimeRange[]>(() =>
    (user?.screen_time ?? []).map((slot) => ({
      id: slot.id,
      startMinutes: parseTimeToMinutes(slot.start_time),
      endMinutes: parseTimeToMinutes(slot.end_time),
    })),
  );

  // Daily cap
  const dailySeconds = user?.daily_screen_seconds ?? 0;
  const [hours, setHours] = useState(
    dailySeconds > 0 ? Math.floor(dailySeconds / 3600) : 1,
  );
  const [minutes, setMinutes] = useState(
    dailySeconds > 0 ? Math.floor((dailySeconds % 3600) / 60) : 0,
  );

  // Saving state
  const [isSaving, setIsSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Whitelist (subscriptions)
  const [channels, setChannels] = useState<GetSubscriptions200ItemsItem[]>([]);
  const [channelSearch, setChannelSearch] = useState("");

  useEffect(() => {
    const { getSubscriptions } = getChannel();
    getSubscriptions({ limit: 50 })
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
    try {
      const { patchUsersMeStatus } = getUser();
      await patchUsersMeStatus({
        screen_time: timeRanges.map((r) => ({
          id: r.id,
          start_time: formatTime(r.startMinutes),
          end_time: formatTime(r.endMinutes),
        })),
        daily_screen_seconds: hours * 3600 + minutes * 60,
      });
      await refreshUser();
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch {
      // Error handling
    } finally {
      setIsSaving(false);
    }
  };

  const clampHours = (v: number) =>
    setHours(Math.max(0, Math.min(23, isNaN(v) ? 0 : v)));
  const clampMinutes = (v: number) =>
    setMinutes(Math.max(0, Math.min(59, isNaN(v) ? 0 : v)));

  const isTimeInvalid = hours === 0 && minutes === 0;

  const handleUnsubscribe = async (channelId: string) => {
    try {
      const { deleteSubscriptionsChannelId } = getChannel();
      await deleteSubscriptionsChannelId(channelId);
      setChannels(channels.filter((ch) => ch.channel_id !== channelId));
    } catch {
      // Error handling
    }
  };

  return (
    <>
      <div class="flex flex-col gap-2 mb-2">
        <h1 class="text-3xl lg:text-4xl font-black leading-tight tracking-[-0.033em]">
          {t("restrictions.title")}
        </h1>
      </div>

      <div class="flex flex-col gap-6">
          {/* Time Constraints */}
          <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
            <div class="p-6 border-b border-border-light dark:border-border-dark">
              <h2 class="text-xl font-bold flex items-center gap-2">
                <span class="material-symbols-outlined text-primary">
                  schedule
                </span>
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
                  <span class="material-symbols-outlined group-hover:scale-110 transition-transform">
                    add
                  </span>
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
                  <div class="relative">
                    <input
                      type="number"
                      class="w-24 pl-4 pr-8 py-3 bg-background-light dark:bg-background-dark border border-border-light dark:border-border-dark rounded-lg focus:ring-2 focus:ring-primary outline-none text-center font-bold text-lg"
                      min={0}
                      max={23}
                      value={hours}
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
                  <button
                    class="ml-auto px-6 py-3 bg-primary hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed text-white font-bold rounded-lg transition-colors shadow-lg shadow-primary/20 flex items-center gap-2 cursor-pointer border-none"
                    disabled={isSaving || isTimeInvalid}
                    onClick={handleSave}
                  >
                    <span class="material-symbols-outlined text-[20px]">
                      save
                    </span>
                    {isSaving ? t("restrictions.saving") : t("restrictions.save")}
                  </button>
                </div>
                {saveSuccess && (
                  <span class="text-sm text-green-600 dark:text-green-400 font-medium flex items-center gap-1">
                    <span class="material-symbols-outlined text-[18px]">
                      check_circle
                    </span>
                    {t("restrictions.saved")}
                  </span>
                )}
                {isTimeInvalid && (
                  <p class="text-sm text-red-500">
                    {t("restrictions.zeroTimeError")}
                  </p>
                )}
              </div>
            </div>
          </div>
          {/* Whitelist */}
          <div class="flex flex-col rounded-xl bg-card-light dark:bg-card-dark shadow-sm border border-border-light dark:border-border-dark overflow-hidden">
            <div class="p-6 border-b border-border-light dark:border-border-dark">
              <h2 class="text-xl font-bold flex items-center gap-2">
                <span class="material-symbols-outlined text-primary">
                  playlist_add_check
                </span>
                {t("restrictions.whitelist")}
              </h2>
              <p class="text-sm text-text-muted-light dark:text-text-muted-dark mt-1">
                {t("restrictions.whitelistDesc")}
              </p>
            </div>
            <div class="p-6 flex flex-col gap-6 grow">
              <div class="relative group">
                <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                  <span class="material-symbols-outlined text-text-muted-light dark:text-text-muted-dark group-focus-within:text-primary transition-colors">
                    search
                  </span>
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
                    <div
                      class="bg-center bg-no-repeat bg-cover rounded-full size-12 shrink-0 border border-border-light dark:border-border-dark"
                      style={{
                        backgroundImage: `url("${ch.external_channel_icon_url}")`,
                      }}
                    />
                    <div class="flex flex-col grow min-w-0">
                      <p class="font-bold truncate">
                        {ch.external_channel_display_name}
                      </p>
                      <p class="text-xs text-text-muted-light dark:text-text-muted-dark">
                        @{ch.channel_custom_id}
                      </p>
                    </div>
                    <button
                      class="size-8 flex items-center justify-center rounded-full text-text-muted-light dark:text-text-muted-dark hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 transition-all cursor-pointer bg-transparent border-none"
                      onClick={() => handleUnsubscribe(ch.channel_id)}
                    >
                      <span class="material-symbols-outlined text-[20px]">
                        close
                      </span>
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>

      </div>
    </>
  );
}
