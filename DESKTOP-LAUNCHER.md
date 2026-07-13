# Making an App Feel Like a Real Desktop App (Optional)

By default, any app in this repo just opens as a tab in your browser.
That's completely fine, everything works. But if you'd rather it open
in its own window, with no address bar or tabs, like a real installed
program, here's how, using a browser you already have.

This takes about two minutes and one small edit.

## What you'll need

Any one of these, already installed (you almost certainly have at
least one): **Brave**, **Google Chrome**, or **Microsoft Edge**.

Don't use a "Nightly," "Canary," "Dev," or "Beta" version of any of
these, those are testing builds, not the one you use every day. If
you're not sure, you're using the regular one.

## Steps

1. Right-click anywhere on your Desktop, choose **New > Shortcut**.
2. Windows will ask for a location. Paste one of the lines below,
   depending on which browser you're using, then click **Next**.

**Brave:**
```
"C:\Program Files\BraveSoftware\Brave-Browser\Application\brave.exe" --app="file:///C:/Users/YOUR-USERNAME/Desktop/AppFolder/appname.html"
```

**Chrome:**
```
"C:\Program Files\Google\Chrome\Application\chrome.exe" --app="file:///C:/Users/YOUR-USERNAME/Desktop/AppFolder/appname.html"
```

**Edge:**
```
"C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe" --app="file:///C:/Users/YOUR-USERNAME/Desktop/AppFolder/appname.html"
```

3. **The one edit you need to make**: replace `YOUR-USERNAME` with
   your actual Windows account name. To find it: open File Explorer,
   click **This PC**, open your **C: drive**, open the **Users**
   folder, your account name is whichever folder there is yours
   (usually your first name).
4. Also replace `AppFolder/appname.html` with wherever you actually
   put the app after unzipping it, and its real filename. Each app's
   own instructions tell you exactly what to put here.
5. Give the shortcut a name, click **Finish**. Double-click it, the
   app opens in its own clean window.

## Optional: give it a proper icon

Right-click the new shortcut, **Properties**, **Change Icon**, browse
to the app's own `icon.ico` file if one is included in its download.
Not required, just a nice touch.

## Optional: extra isolation, no new browser needed

If you'd rather this app's saved data live somewhere completely
separate from your everyday browsing (a dedicated, isolated profile),
you don't need to install a separate browser channel to get that.
Just add one more piece to the same shortcut target from step 2:

```
--user-data-dir="C:\Users\YOUR-USERNAME\AppData\Local\AppProfiles\appname"
```

Add it right after the `--app=` part, same line. First launch creates
that folder automatically. This gives full separation from your main
browser profile using the exact same browser you already have, no
extra install, no update-stability risk from running a beta channel.

## A note for Firefox users

This specific trick is a Chromium feature (Brave, Chrome, and Edge are
all built on Chromium), Firefox doesn't have a matching version of it.
The app works exactly the same in Firefox either way, you'll just see
the normal browser window with the address bar, rather than a
toolbar-free app window.
