Below is a practical, step-by-step design for turning your web-prototype timer into a polished **iPhone app built with Expo**.  Everything is organised so you can start a new expo project, paste code, and iterate quickly—even if you’ve never shipped a mobile app before.

---

## 1. Project bootstrap

```bash
# Node ≥18
npm install --global expo-cli
expo init film-dev-timer --npm --template expo-template-blank-typescript
cd film-dev-timer
```

Add must-have libraries:

| Purpose           | Library                                                                           |
| ----------------- | --------------------------------------------------------------------------------- |
| Navigation        | `@react-navigation/native`, `@react-navigation/stack`                             |
| Sounds / haptics  | `expo-av`, `expo-haptics`                                                         |
| Keep screen awake | `expo-keep-awake`                                                                 |
| Local DB          | `expo-sqlite` (or start with `@react-native-async-storage/async-storage`)         |
| Tailwind-in-RN    | `nativewind` (optional – lets you keep the Tailwind-like classes you already use) |

```bash
expo install expo-av expo-haptics expo-keep-awake expo-sqlite
npm i nativewind
```

---

## 2. Directory layout

```
/src
  ├── components/
  │     TimerDisplay.tsx
  │     ProcessStepCard.tsx
  ├── data/
  │     rolls.json          ← static seed data
  │     steps.json
  ├── db/
  │     schema.ts
  │     useRolls.ts         ← thin SQLite/AsyncStorage wrapper
  ├── screens/
  │     HomeScreen.tsx      ← roll selector
  │     TimerScreen.tsx     ← main UI
  │     SettingsScreen.tsx
  ├── providers/
  │     RollsProvider.tsx   ← React Context over DB
  └── navigation/
        index.tsx           ← Stack & Tab config
```

Separation keeps the timer logic portable and makes it easier to unit-test.

---

## 3. Data model & storage

### a. JSON seed files (shipped with the app)

```jsonc
// rolls.json (abridged)
[
  { "id": 1, "film": "120", "pushPull": "+1", "devSec": 295, "notes": "Fresh" },
  ...
]

// steps.json
[
  { "id": "presoak",   "label": "Pre-soak",  "sec": 60 },
  { "id": "developer", "label": "Developer", "sec": null /* varies */ },
  { "id": "blix",      "label": "Bleach-fix", "sec": 390 },
  ...
]
```

### b. Persistent layer

| Option                            | Pros                               | Cons                                | Suggested use                                               |
| --------------------------------- | ---------------------------------- | ----------------------------------- | ----------------------------------------------------------- |
| **AsyncStorage** (key-value JSON) | Dead simple                        | No SQL / queries; 100 kB soft limit | Great for v1 if you only write 🤏 config & completion flags |
| **SQLite** (`expo-sqlite`)        | True relational db, ACID,  \~50 MB | Slightly more code                  | Use when you start logging every roll & exporting CSVs      |

The provider hook (`useRolls.ts`) owns:

* `getRolls()` – load from seed → override with DB values
* `markStepComplete(rollId, stepId, ts)`
* `saveSetting(key, value)`

Stick with TypeScript types so the compiler nags you when schema changes.

---

## 4. Navigation shell

```tsx
export default function App() {
  return (
    <RollsProvider>
      <NavigationContainer theme={DarkTheme}>
        <Stack.Navigator screenOptions={{ headerShown: false }}>
          <Stack.Screen name="Home" component={HomeScreen} />
          <Stack.Screen name="Timer" component={TimerScreen} />
          <Stack.Screen name="Settings" component={SettingsScreen} />
        </Stack.Navigator>
      </NavigationContainer>
    </RollsProvider>
  );
}
```

DarkTheme matches your existing black-background design—ideal for a dim darkroom.

---

## 5. Porting the Web prototype to React Native

### a. Replace DOM elements

| Web                        | React Native                                          |
| -------------------------- | ----------------------------------------------------- |
| `<div>`                    | `<View>`                                              |
| `<span>` / text in buttons | `<Text>`                                              |
| CSS / Tailwind classes     | `className` via **nativewind** or `StyleSheet.create` |
| `<audio>` tag              | `Audio.Sound` from **expo-av**                        |

The logical structure, state variables (`currentRoll`, `currentStep`, `timeLeft` …) and timer algorithm from the prototype  stay 99 % identical—only the rendering changes.

### b. Timer logic adjustments

* Use `useEffect` + `setInterval` exactly as in web code (works the same).
* Pair audio cue with **Haptics.notificationAsync** so you can “feel” the alert if phone is in pocket.
* Call **activateKeepAwake()** from `expo-keep-awake` while a timer is running so iOS won’t lock the screen halfway through development.

### c. Background notifications (optional but nice)

When the user switches apps you can schedule a one-shot push:

```ts
import * as Notifications from 'expo-notifications';

await Notifications.scheduleNotificationAsync({
  content: { title: 'Developer step finished', sound: true },
  trigger: { seconds: remainingSec }
});
```

Apple kills JS timers in background; this ensures the alert still fires.

---

## 6. UI polish for iPhone

| Guideline             | Implementation                                             |
| --------------------- | ---------------------------------------------------------- |
| **Safe area**         | Wrap screens in `<SafeAreaView>`                           |
| **One-hand reach**    | Main control buttons at bottom (use absolute `bottom: 32`) |
| **Big targets**       | 48 px min button height (already true)                     |
| **Dynamic type**      | Scale font sizes with `useWindowDimensions()`              |
| **System status bar** | `expo-status-bar` with a dim red tint (`style="light"`)    |
| **Offline-first**     | Everything local; no remote calls needed                   |
| **Short sounds**      | 0.2 s WAV in `/assets/sounds/notify.wav`; load on demand   |

---

## 7. Replenishment helper

Show a modal sheet when the stabiliser step completes. Drive the exact copy from **steps.json** so translations or chemistry variants can swap text without editing code .

---

## 8. Build & ship

1. **Convert to EAS** (Expo Application Services):
   `eas build:configure`
2. **Create an Apple “App ID”** and App Store entry.
3. `eas build -p ios --profile preview` → generates an `.ipa` in the cloud.
4. `eas submit -p ios` to TestFlight.

You can stay on the free Expo Go app for local dev; deep links `exp://` reload instantly.

---

## 9. Future-proofing

| Next feature       | Why               | Library                                    |
| ------------------ | ----------------- | ------------------------------------------ |
| Export CSV log     | Archival          | `expo-sharing` + `react-native-csv`        |
| iCloud backup      | Cross-device      | `expo-file-system` + iCloud container path |
| Apple Watch glance | Hands-free timing | `react-native-watch-connectivity`          |
| In-app presets     | B\&W, E-6         | param files in `/src/data/`                |

---

### TL;DR game plan

1. **Expo init**, add navigation, SQLite (or AsyncStorage), AV, haptics.
2. **Copy logic** from your prototype almost unchanged; swap web tags for RN ones.
3. **Persist** rolls & steps via a tiny data layer; seed with JSON.
4. Dark, big, simple UI + keep-awake + audio/haptic alert = perfect darkroom companion.
5. Ship through **EAS Build** to TestFlight; iterate with OTA updates.

With this structure you can get an MVP running on your iPhone in an evening and refine from there. Happy (film) developing!
