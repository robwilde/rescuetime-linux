# Rescue Time activity recorder

I would like you to use the information outlined in the `context/rescuetime/api-docs.md` file, specifically the bottom section title `Documentation for the Offline Time POST API`
to record what app I am using on this system and for how long. Then every 15 minutes, send that info to rescuetime. 

## Details

This is an example of the information you captured in the test

```shell
Active Window: can-eye-budget – README.md (jetbrains-phpstorm) [00:03:30]
```

This is perfect info as it has the application_name `(jetbrains-phpstorm)`, the time `[00:03:30]` as well as the project information and file name `can-eye-budget – README.md`.

You can use this info to create the payload data. Use the application_name for the activity_name and, in this case project information and file as the activity details.

```json
{  "start_time": "2020-01-01 09:00:00",  "duration": 60,  "activity_name": "Meeting",  "activity_details": "Daily Planning"  }
```

Another good example is when I access my emails in Wavebox.

```shell
Active Window: Inbox (1,064) - robert@mrwilde.com - MrWilde Mail - Wavebox (wavebox) [00:04:42]
```

Using the information in the `()` that is the application_name as the activity and the other information as the description.

You can log the activity and then use the saved `[]` time_data to calculate how much time was spent in each application.