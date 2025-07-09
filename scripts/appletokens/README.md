# Apple Token Refresh Script

Every 6 months the Apple Developer Token expires and so with it the Apple Music User token. This script and HTML file serve as a way to refresh these tokens.

1. Go to [Apple Developer Page for Keys](https://developer.apple.com/account/resources/authkeys/list) and delete the old key.
2. Create a new key with MusicKit enabled under media services.
3. Save the downloaded key into this directory as a `key.p8` file.
4. Run the script with the following environment variables: `TEAM_ID` and `KEY_ID`. `KEY_ID` can be viewed when looking at the view key details page for the key we generated in step 2. `TEAM_ID` should be visible in the top right of the page, and is the ID for the Apple Developer Account.

```bash
TEAM_ID="..." KEY_ID="..." go run main.go
```

5. Copy and save the given JWT that is outputted from the script. This is our **DEVELOPER TOKEN**.
6. To get the user auth token replace the `DEVELOPER TOKEN` in [auth.html](./auth.html) with the given developer token from step 5. This webpage needs to be visited from `http` and not `file://`. This can be done using the "Show Preview" command when right clicking on the file in vscode.
7. Copy the user token that is outputted on the web page.
8. Update the `APPLE_MUSIC_APP_TOKEN` to the developer token from step 4.
9. Update the `APPLE_MUSIC_USER_TOKEN` to the user token from step 6.
