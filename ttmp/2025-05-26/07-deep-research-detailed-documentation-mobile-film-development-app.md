Thanks! I’ll prepare a comprehensive implementation plan and documentation tailored for an intern to build this Expo-based iPhone film development timer app. It will cover:

* Full setup with Expo and SQLite
* Local development flow on iPhone
* Expo-compatible background notifications
* UI design patterns for mobile
* Deployment to TestFlight for App Store distribution
* Inline documentation for all relevant libraries

I’ll get started and let you know as soon as it’s ready.


# Mobile Film Development Timer App – Implementation Plan

## Overview

This project is a **mobile film development timer** app for iPhone, built with **React Native** using the **Expo** framework. The app will let users time multiple steps of film developing (e.g. developer bath, stop bath, fixer) entirely offline. Key features include:

* **Persistent Offline Storage:** using a local **SQLite** database (via `expo-sqlite`) to store preset development processes and logs, so no network or cloud sync is needed.
* **Local Notifications for Timers:** scheduling **background notifications** (via `expo-notifications`) to alert the user when each step’s timer finishes – even if the app is closed or in the background.
* **iOS-Native Design:** a user interface that follows iOS design guidelines (proper layout for different iPhone screens, safe areas, typical iOS navigation patterns) to feel like a true native app.
* **Expo Workflow:** development via Expo for quick iteration and easy preview on a physical iPhone, and deployment using **Expo Application Services (EAS)** for TestFlight and App Store distribution.

Throughout this guide, we provide a step-by-step plan with **sample code** snippets, project structure suggestions, and links to relevant documentation for further reference. The target audience is an **intern-level developer**, so we will explain concepts clearly and keep the implementation straightforward.

## Project Setup and Development Environment

To start, ensure the development environment is ready for React Native with Expo:

1. **Install Node.js and Expo CLI:** You should have Node.js installed. Install Expo’s CLI tools by running `npm install --global expo-cli` (or use `npx` as needed).

2. **Create a New Expo Project:** Use the Expo CLI to bootstrap a project. For example, in your terminal run:

   ```bash
   npx create-expo-app film-timer
   ```

   This will create a new React Native project with Expo. Choose the **blank template** (JavaScript or TypeScript as you prefer).

3. **Add Required Expo Modules:** Install the libraries for SQLite and notifications:

   ```bash
   npx expo install expo-sqlite expo-notifications
   ```

   Expo SDK will auto-link these. The `expo-sqlite` module provides a persistent SQLite database that remains intact across app restarts. The `expo-notifications` module will let us schedule local notifications (in-app alerts) without any server, which Expo supports out of the box.

4. **iOS Permissions Configuration:** Since we plan to use notifications, update the Expo app configuration to request iOS notification permissions. In the `app.json` (or `app.config.js`), add an iOS section with `"ios": { "usesAppleSignIn": false, "infoPlist": { "NSUserNotificationUsageDescription": "We use notifications to remind you when film development steps are done." } }`. This ensures iOS knows *why* the app will send notifications (required for approval).

5. **Run the Project Locally:** Start the development server with `expo start`. You can then preview on an iPhone:

   * **Using Expo Go:** After running the start command, a QR code appears in the terminal. Scan the QR code with your iPhone’s Camera app to launch the project in the Expo Go app. Ensure the phone and computer are on the same network (or use “Tunnel” mode if needed for remote networks). This lets you see live updates on the device as you save code changes.
   * **Using iOS Simulator (Optional):** If you have a Mac with Xcode, you can press `i` in the Expo CLI to open the iOS Simulator. Keep in mind that notifications do **not** work on the iOS simulator (you need a real device for testing notifications).

With the project running on a device, you should see the default Expo welcome screen. From here, you can start implementing features.

## Data Persistence with SQLite (Offline Storage)

For offline functionality, we will use **expo-sqlite** to store data on the device. SQLite allows structured storage and querying of data with SQL, perfect for saving film development **recipes** (each recipe being a series of timed steps) and past results. The database will persist between app launches (the data stays on the phone).

**Setting up the Database:** You can open a SQLite database (it will be created if it doesn’t exist). For example, initialize a database in your app’s code (e.g. in an initialization module or on app load):

```javascript
import * as SQLite from 'expo-sqlite';

const db = SQLite.openDatabase('filmDevTimer.db');  // creates or opens the database file

// Create tables if they don't exist
db.transaction(tx => {
  tx.executeSql(
    `CREATE TABLE IF NOT EXISTS recipes (
       id INTEGER PRIMARY KEY AUTOINCREMENT, 
       name TEXT
     );`
  );
  tx.executeSql(
    `CREATE TABLE IF NOT EXISTS steps (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       recipe_id INTEGER,
       step_name TEXT, 
       duration INTEGER, 
       FOREIGN KEY(recipe_id) REFERENCES recipes(id)
     );`
  );
});
```

*The code above opens (or creates) a database and sets up two tables: one for film development recipes and one for the steps in each recipe. We use `tx.executeSql` within a transaction to execute SQL statements.*

In the schema above, a **recipe** might be something like "Kodak T-Max 400 at 20°C", and the **steps** table stores entries like “Developer – 7 minutes”, “Stop – 1 minute”, “Fixer – 5 minutes” linked by `recipe_id`. You can adjust the schema to your needs (for example, add columns for temperature or chemical dilutions if needed).

**Inserting and Querying Data:** You can insert new data and fetch data similarly using `executeSql`. For example, to insert a new step into a recipe:

```javascript
function addStep(recipeId, name, durationSeconds) {
  db.transaction(tx => {
    tx.executeSql(
      'INSERT INTO steps (recipe_id, step_name, duration) VALUES (?, ?, ?);',
      [recipeId, name, durationSeconds],
      (_, result) => console.log('Step added with ID:', result.insertId),
      (_, error) => console.error('Error inserting step:', error)
    );
  });
}
```

To query data (e.g. get all steps for a recipe), you would use `tx.executeSql` and read from the `rows` returned:

```javascript
tx.executeSql(
  'SELECT * FROM steps WHERE recipe_id = ? ORDER BY id;',
  [recipeId],
  (_,{ rows }) => {
    const steps = rows._array;  // array of step objects
    console.log('Loaded steps:', steps);
  }
);
```

Using SQLite has the advantage that complex queries (e.g. filtering, ordering) can be done in SQL. It’s efficient and works fully offline. For an **intern**-level developer, it may help to use a small wrapper or ORM, but for this app direct SQL usage is fine. Refer to the Expo SQLite docs for more advanced usage patterns (like using `executeSql` promise wrappers or transactions).

**Note:** While developing, you can use tools like Expo’s SQLite inspector or export the database file for debugging, but for our purposes, logging query results to console (visible in Expo debug console) is usually enough to verify it’s working.

## Implementing Timer Logic and Background Notifications

The core of the app is timing each step and alerting the user when a step is complete. We’ll combine React Native’s timing functions with Expo’s local notifications to achieve this.

### Timer Countdown Implementation

Inside the app, you might have a screen that displays the current step’s remaining time counting down. There are a few ways to implement a countdown timer in React Native:

* Using the built-in `setInterval` or `setTimeout` to update a counter in state every second.
* Using a library or hook (like `useEffect` with a timer) to decrease the time.

For simplicity, you can use a `useEffect` with `setInterval`. For example, if you have a component for the active step timer:

```javascript
const [secondsLeft, setSecondsLeft] = useState(step.duration); 

useEffect(() => {
  if (secondsLeft <= 0) return;
  const interval = setInterval(() => {
    setSecondsLeft(sec => sec - 1);
  }, 1000);
  return () => clearInterval(interval);
}, [secondsLeft]);
```

This will update `secondsLeft` state every second. You would also display `secondsLeft` in your UI, formatted as mm\:ss. When `secondsLeft` hits 0, that step is done – at that point you’d trigger moving to the next step and/or notifying the user.

However, **remember**: if the app goes into background or is closed, JavaScript timers *will not run*. This is why we need to schedule local notifications with the system to handle alerts when the app isn’t active. Essentially, we pre-schedule the notifications for each step as soon as the timer starts.

### Scheduling Local Notifications

Using Expo’s Notifications API, we can schedule notifications to fire at specific times. A **local notification** is one scheduled by the app itself (no server needed) to appear after some delay or at a specific time. For our film timer, when the user starts a recipe, we will schedule a notification for the end of each step. That way, even if the user switches apps or locks the phone (or the app is terminated), the phone will still show the alert at the right time.

**Setup Notifications:** First, request the user’s permission to send notifications (usually on first app use). You should do this early, perhaps when the app launches or when starting the first timer:

```javascript
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

// Request permissions (iOS will prompt the user)
async function requestNotificationPermission() {
  const { status } = await Notifications.requestPermissionsAsync();
  if (status !== 'granted') {
    alert('Please enable notifications to get timer alerts.');
  }
}
```

You might call `requestNotificationPermission()` on app startup or when the user starts a timer sequence, to ensure you can schedule notifications.

Also, configure the notification handling behavior (to define how notifications are shown when the app is foreground). Expo provides a default handler; we can use it to automatically show a banner if the app is open:

```javascript
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
  }),
});
```

This ensures that if a notification fires while the app is in foreground, an alert is still shown (by default, iOS would not show a notification banner if the app is active).

**Schedule step notifications:** Now, when the user starts the timer for a recipe, loop through each step and schedule a notification for the end of that step. For example:

```javascript
// Assume steps is an array of step objects with { step_name, duration } in seconds
async function startRecipeTimer(steps) {
  let accumulatedTime = 0;
  for (const [index, step] of steps.entries()) {
    accumulatedTime += step.duration;  // seconds from start until end of this step
    // Schedule notification at the time this step finishes:
    await Notifications.scheduleNotificationAsync({
      content: {
        title: "Step Complete",
        body: `The "${step.step_name}" step is done! ${index < steps.length - 1 ? 'Start the next step.' : 'All steps completed.'}`,
        sound: true,  // play default sound (you can customize in app.json if desired)
      },
      trigger: { seconds: accumulatedTime }  // fire after accumulatedTime seconds from now
    });
  }
}
```

In this snippet, we schedule a notification for each step’s completion time. For instance, if step 1 is 300s (5 minutes) and step 2 is 60s (1 minute), we schedule one notification at 300s, and another at 360s from now, and so on. The content of the notification includes the step name and a prompt to move to the next step (or a completion message if it was the last step).

> **Tip:** Expo’s scheduling can accept either a relative time interval (as shown above) or an absolute timestamp (specific date/time) for the trigger. Using a relative `seconds` offset is straightforward here. If `trigger: null` is used, the notification would fire immediately.

The function returns a notification **identifier** for each scheduled notification (we awaited them above). If needed, you can store these identifiers (e.g. in state or a ref) in case you want to cancel notifications (Expo provides a `Notifications.cancelScheduledNotificationAsync(id)` to cancel by identifier). Cancellation might be useful if the user aborts the timer midway or wants to reschedule.

**User Experience of Notifications:** When a notification triggers on iOS, the user will hear the default sound (unless silenced) and see the notification message, even if the app is backgrounded or terminated. Tapping the notification can be configured to open the app and possibly navigate to a relevant screen. By default, tapping it will just open the app; you can handle deeper linking if desired (e.g., via a listener in Expo notifications to react when a notification is selected).

Expo’s local notifications are reliable for offline alerts and do not require any push setup or servers. According to Expo docs, notifications will alert the user even if the app isn’t active – exactly what we need for a timer. Just make sure the app requests permission and schedules them properly before going to background.

**Important:** During development, test notifications on a real iPhone. As noted earlier, the iOS Simulator does not support incoming local notifications. Also, if you’re using Expo Go in development, local scheduled notifications should work on device (Expo Go supports scheduling local notifications, whereas push notifications require a custom dev build). When testing, try starting a timer, then closing the app (swipe it away) to verify that the notifications still fire at the expected times. (They should, because the scheduling is handled by iOS once set, even though your JS app isn’t running).

### Handling Timer Completion and Next Steps

In the app’s UI, when each step completes, you’ll likely want to automatically move to the next timer or prompt the user. You can handle this by listening for the notification while the app is foregrounded, or simply by checking if the app is still open when the timer ends.

One approach: maintain an index of the current step in state, and when `secondsLeft` hits 0 for a step, increment to the next step and reset the timer state. If the app is backgrounded, the notification will bring the user’s attention back, and you can design the app to show the current step or a completion summary when reopened.

For completeness, you can add a listener for notifications when the app is running. For example, Expo Notifications can add a listener via `Notifications.addNotificationResponseReceivedListener` to handle the user tapping a notification (to maybe navigate to the relevant screen in your app). This is an advanced enhancement – for an intern-level first version, it can be skipped or kept simple (the default behavior of opening the app is usually fine).

## UI Design and iOS-Specific Considerations

Design the user interface to feel at home on iOS. Here are some guidelines and tips:

* **Use Native Components:** React Native core components (like `Button`, `TextInput`, `Picker`, etc.) will automatically render in the iOS style by default. Use these standard components for familiarity. For navigation, you can use a library like React Navigation (which by default on iOS uses native-like slide transitions and header styles) or Expo Router. Keep the navigation simple – perhaps a home screen listing recipes and a timer screen for the active process.
* **Follow Human Interface Guidelines (HIG):** Strive for a clean, uncluttered interface with easy-to-read typography. Apple’s HIG recommends a minimum tap target size (44px), ample whitespace, and using the system font for readability. Use **Safe Areas** to ensure content isn’t hidden by notches or the iOS status bar. For example, wrap your top-level view in `<SafeAreaView style={{flex: 1}}> ... </SafeAreaView>` (from `react-native-safe-area-context`, which Expo includes by default) so that on devices like iPhone 14 or those with a notch, your content (e.g. headers or bottom buttons) aren’t cut off by the curved edges, status bar, or home indicator.
* **iOS Aesthetics:** Use iOS-style design elements where possible. For instance, iOS typically uses segmented controls or pickers. If your app lets users input times or choose chemical processes, consider using the iOS date/time picker or a wheel picker style for a native feel. Expo includes community modules like `@react-native-community/datetimepicker` if needed (you can install if required). Even without additional libraries, a simple ScrollView and Picker can be styled to mimic native selectors.
* **Consistent Navigation:** If your app has multiple screens, ensure there’s a clear way to navigate (e.g., a back button on top-left if using a stack navigator). On iOS, the gesture to swipe back (if using React Navigation stack) works automatically. If you use a tab bar for different sections of the app, iOS places the tab bar at the bottom with icons and labels – you can use React Navigation’s bottom tab navigator which follows the native convention.
* **Feedback and Haptics:** iPhones have taptic engines. Using subtle haptic feedback can enhance the experience (Expo has `Haptics` module for light taps). For example, when a timer step ends or when the user presses the start button, you could trigger a light vibration. This is optional, but it’s a nice iOS touch.
* **Theming and Dark Mode:** By default, Expo apps use light mode. You can support Dark Mode (iOS users appreciate this) by using `Appearance` from React Native to detect theme, or simply ensure your color choices adapt (e.g., mostly using system controls or black text on white which automatically invert in dark mode if using the system components). This can be an enhancement after basic functionality.

Remember to test the UI on an actual iPhone. Look at it on different screen sizes if possible (e.g., iPhone SE vs. iPhone 14 Pro Max) to ensure layout scales nicely. Expo’s responsive flexbox layout will handle most sizing, but be mindful of any absolute positioning.

**Resources:** Apple’s Human Interface Guidelines document is a great resource for design principles. In summary, aim for **clarity, deference, and depth** – meaning the design should be clear and legible, not distracting, and use hierarchy to show what’s important. Keep controls familiar (an intern can look at common iOS timer apps for inspiration, such as the Clock app’s timer or existing film timer apps).

## Testing on a Physical iPhone (Local Development)

During development, regularly test the app on a physical iPhone to catch platform-specific issues and to experience the app as users will. Using **Expo Go** makes this very easy:

* **Running in Expo Go:** As described earlier, after running `expo start`, scan the QR code with your iPhone camera to open the project. This will load your JavaScript bundle in the Expo Go app on the device. Every time you save changes in your code, the app will hot-reload on the phone. This lets you rapidly iterate on UI and functionality using the phone’s screen and sensors. *Tip: if the phone cannot reach your dev server via LAN, use `expo start --tunnel` to route through the internet (slower but works on any network).*

* **Debugging:** You can use `console.log` and see logs in the Expo CLI or in the browser Developer Tools (press `d` in the terminal to get a debug URL). This is helpful for seeing output from SQLite queries or timing events. For a more interactive debug, you can use Expo’s support for React DevTools or breakpoints, but console logs are usually sufficient at this stage.

* **Testing Offline:** Since the app is meant to work offline, try running it with no internet connection on the device. Expo Go itself needs a connection to load the bundle initially, but once loaded, your app’s features (timers, local DB, local notifications) should not require internet. Verify that you can start a timer in Airplane Mode and still get the notification at the end of the step (you should, because everything is local).

* **Testing Edge Cases:** Try closing the app while a timer is running (swipe it closed) and confirm the notification still appears at the right time – this validates the background notification functionality. Also test what happens if a user starts a timer and then starts another without finishing the first (you may need to handle canceling prior notifications in such a case to avoid duplicate alerts). These help iron out logic issues.

* **Iteration:** Encourage quick iteration: since this is an intern project, making incremental changes and seeing results on the phone will build understanding. Expo’s fast refresh will help the intern tweak UI layouts, fix any runtime errors, and test on the device frequently.

## Preparing for TestFlight and App Store Deployment (using EAS)

Once the app is functioning and tested in development, the next step is to distribute it via Apple’s TestFlight and ultimately the App Store. Expo Application Services (EAS) will assist in building the app binary and uploading it. Below is a plan to get the app from development to TestFlight:

1. **Apple Developer Account:** Ensure you have access to an Apple Developer account (enroll as an Apple Developer which requires a yearly fee). You’ll use this to manage App Store Connect and certificates. Also, decide on a unique **Bundle Identifier** for your app (e.g. `"com.yourcompany.filmtimer"`). Set this in your `app.json` under the `ios.bundleIdentifier` field before building.

2. **EAS Build Configuration:** Expo uses **EAS Build** to compile your app in the cloud. EAS requires an Expo account (run `expo login` if you haven’t). In your project, create an `eas.json` (if not already present) which holds build profiles. The default should have a `"production"` profile suitable for App Store. For example:

   ```json
   {
     "build": {
       "production": {}
     }
   }
   ```

   (Expo may have initialized this for you). The production profile will produce an App Store-ready build (an **IPA** file).

3. **Build for iOS (Archive):** Run the build command for iOS:

   ```bash
   eas build --platform ios --profile production
   ```

   The EAS CLI will guide you through the process. The first time, it will ask to authenticate with Apple and to create or reuse signing credentials (distribution certificate and provisioning profile). You can generally let Expo handle these automatically – just follow the prompts and enter your Apple Developer Apple ID when asked. EAS will create a **production build** of your app, optimized for TestFlight/App Store distribution. This might take a few minutes on Expo’s servers. Once complete, you’ll get a URL to the build artifact or you can find it on your Expo dashboard.

4. **Submit to TestFlight:** After a successful build, you need to upload it to Apple’s App Store Connect. Expo provides **EAS Submit** to automate this. Run:

   ```bash
   eas submit --platform ios --latest
   ```

   This will take the latest build (from step 3) and submit it to App Store Connect. You’ll likely be prompted for an **Apple App-Specific Password** (which you generate from your Apple ID account for third-party upload tools). Provide that when asked. The CLI will then use Apple’s API (via Fastlane internally) to upload the IPA.
   **Alternatively:** You can manually upload the IPA through Xcode’s Transporter or App Store Connect web, but EAS Submit is simpler.

5. **App Store Connect – TestFlight:** Once uploaded, log in to [App Store Connect](https://appstoreconnect.apple.com/) and find your app (you might need to create a new app entry first if this is the first upload – include name, bundle ID, etc., matching what you set in app.json). In the app’s page, go to **TestFlight** tab. You should see the build you uploaded (it may take a few minutes to process). Enable TestFlight testing for internal testers:

   * Add yourself (and any team members or the intern’s Apple ID) as an **Internal Tester** (up to 100 emails can be added without review). This is under “Users and Roles” -> TestFlight, or within the TestFlight section of the app.
   * Create a testing group and add testers, then select the build and start testing. Apple will send an email invite to testers.
   * On your iPhone, install the **TestFlight app** from the App Store. The invite email will have a redeem code or a link. Accept the invite, and you’ll see the app in TestFlight. Tap to install it.

   Now you have the app running via TestFlight, independent of Expo Go. TestFlight builds can receive push notifications and other capabilities like a normal app. At this point, test the app thoroughly in this production-like build.

6. **Beta App Review (External TestFlight):** (Optional) If you need to send to external testers (outside your team), you’ll have to submit for Beta App Review on App Store Connect. This involves filling in some compliance info (e.g., encryption usage – for a simple timer, say No to encryption since you’re not using it) and sending to Apple for approval. Since this app is likely internal or for learning, internal testing might suffice.

7. **App Store Submission:** When ready to release to the public, you’ll use App Store Connect to submit the app for review:

   * In App Store Connect, provide all required metadata (screenshots, description, category, privacy policy URL, etc.).
   * Select the build (the same one from TestFlight can be promoted if it’s a version you’re happy with).
   * Submit to App Store Review. Apple will review the app for compliance with guidelines. Since the app is offline and utility-focused, make sure to describe any special features (like why it needs notification access – which should be obvious as a timer).
   * Once approved, you can release the app on the App Store.

Expo’s documentation covers the release process in detail, and it matches the above steps. Essentially, **build** your app, **submit** it to TestFlight, test it, then **submit to App Store** for review. Expo EAS makes the build and submit steps straightforward, automating certificate management and upload. As a reference, the Expo tutorial states: *“we'll create our app's production version and submit it for testing using TestFlight, then submit for App Store review to get it on the App Store”*.

**EAS Build/Submit Documentation:** For further reading, check Expo’s guides on [creating your first build](https://docs.expo.dev/build/setup/) and [submitting to the Apple App Store](https://docs.expo.dev/submit/ios/). They provide step-by-step instructions and troubleshooting tips. For instance, if you encounter issues like the build not appearing in TestFlight, ensure the app’s version and build number in `app.json` match what’s expected in App Store Connect, and check that you created the app listing in App Store Connect if it wasn’t automatically created by EAS Submit.

## Structure and Additional Recommendations

**Project Structure:** Keep the project organized for readability:

* You might have a component structure like `components/` for smaller UI pieces, `screens/` for full-screen components (RecipeListScreen, TimerScreen, etc.), and a `database/` or `services/` folder for SQLite logic (e.g., a module that initializes the DB and functions for CRUD operations). This separation helps an intern see the division between UI and data logic.
* Example file setup:

  * `App.js` (or .tsx) – setup Navigation and global providers (if any).
  * `screens/RecipeListScreen.js` – lists saved recipes from SQLite, allow selecting one to start.
  * `screens/TimerScreen.js` – runs the timers for steps and shows UI, also schedules notifications.
  * `services/database.js` – exports functions to get recipes, add recipe, get steps, add step, using expo-sqlite (using the patterns shown earlier).
  * `services/notifications.js` – perhaps a helper to initialize notifications and schedule them (wrapping the expo-notifications calls).
  * You can also keep things simpler and just do it in components if that’s easier, but showing an organized structure is good practice.

**Relevant Libraries Documentation:** Here are some links for quick reference that you (or the intern) can consult while implementing:

* **Expo SQLite:** Official docs for expo-sqlite explaining the API for executing SQL queries. It includes examples similar to the ones in this guide.
* **Expo Notifications:** Official docs and guides for setting up notifications. Expo’s guide on local notifications vs push is useful. There’s also a “Managing Scheduled Notifications” guide on Medium that can provide a different perspective if needed.
* **React Navigation:** If using navigation, the [React Navigation docs](https://reactnavigation.org/docs/getting-started) will help in setting up a Stack or Tab navigator with Expo.
* **Apple HIG (Human Interface Guidelines):** While not required reading for coding, browsing Apple’s HIG on layout and components can inspire UI decisions. The safe area exposition in Expo docs is a concise reminder of iOS layout basics.

Finally, make sure to maintain **short development cycles**: implement one feature at a time, test on device, and use git for version control to track changes. By following this plan and leveraging the cited resources, an intern developer should be able to build a functioning, polished iPhone app that assists photographers in timing their film development process – all while learning about React Native, Expo, and the iOS app ecosystem. Good luck, and happy coding!

**Sources:**

* Expo Documentation – SQLite (persistent local database)
* Expo Documentation – Notifications (local scheduling and behavior)
* Expo Documentation – Building and Submitting apps with EAS
* Expo Documentation – Running Expo apps on devices (Expo Go usage)
* Expo Documentation – Safe Area and iOS UI considerations
