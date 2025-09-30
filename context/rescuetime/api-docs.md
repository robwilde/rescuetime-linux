This page gives an overview of the RescueTime Developer APIs, details on how to access them, and the types of data that can be retrieved.

User data, particularly activity logs, is synced to the RescueTime servers on a set interval depending on the user's plan subscription. Premium/Organization (paid) plan users' activities are synced every 3 minutes, Lite (free) plan users' activities are synced every 30 minutes. Once the RescueTime app has synced with our servers, the data is immediately available in API results.

## Table of contents

-   [Work/Productivity Equivalents](https://www.rescuetime.com/rtx/settings/api/documentation#work-productivity-equivalents)
-   [Connecting to the RescueTime API](https://www.rescuetime.com/rtx/settings/api/documentation#connection-methods)
-   [API endpoints](https://www.rescuetime.com/rtx/settings/api/documentation#api-endpoints-overview)
    -   [Analytic Data API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#analytic-api-reference)
    -   [Daily Summary Feed API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#daily-summary-feed-reference)
    -   [Alerts Feed API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#alerts-feed-reference)
    -   [Daily Highlights Feed API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#highlights-feed-reference)
    -   [Daily Highlights Post API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#highlights-post-reference)
    -   [Focus Session Trigger API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#focustime-trigger-reference)
    -   [Focus Session Feed API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#focustime-feed-reference)
    -   [Offline Time POST API Documentation](https://www.rescuetime.com/rtx/settings/api/documentation#offline-time-reference)

## Work/Productivity Equivalents

The following table shows the productivity level equivalents referenced in the documentation below for various types of activities.

|<u>Activity</u>|<u>Productivity Equivalent</u>|
|---|---|
|Focus Work|Very Productive|
|Other Work|Productive|
|Neutral|Neutral|
|Personal|Distracting|
|Distracting|Very Distracting|

## Connecting to the RescueTime API

There are two ways to connect to a user’s data in RescueTime, via **API Keys** or via an **Oauth2 Connection**. The API keys are ideal for personal use, while web services wishing to offer a way for their users to use their RescueTime data should use an Oauth2 connection.

### API Key Access:

Users can set up an API key by going to their [key management page](https://www.rescuetime.com/anapi/manage) and creating a new key. That key must be included with each API request. The user can revoke keys at any time.

### Oauth2 Connection Access:

Please [contact us](https://www.rescuetime.com/users/help) and we’ll work with you to set up an Oauth2 application. After the user connects their account via Oauth2, an access token will be created. That access token must be included in each API request. The user can revoke access to a service at any time.

**Oauth2 Authorization Flow**

Once we have created your Oauth2 application, your app will be assigned a `client_id` and a `client_secret`. These tokens can then be used, along with the secure `redirect_uri` provided in the request for an oauth application, in the following steps to authorize your application's user access to their RescueTime data:

-   A `GET` request to the following URL:

    ```
    https://www.rescuetime.com/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=YOUR_REDIRECT_URI&response_type=code&scope=YOUR_SCOPES
    ```

    This will return the `auth_code`, which will expire in 10 minutes.
-   A `POST` request to the following URL:

    ```
    https://www.rescuetime.com/oauth/token
    ```

    with a JSON body of:

    `{  "client_id": YOUR_CLIENT_ID,  "client_secret": YOUR_APP_SECRET,  "grant_type": "authorization_code",  "code": AUTH_CODE,  "redirect_uri": YOUR_REDIRECT_URI  }`

    This will return the `access_token`

-   Data can then be accessed with the `access_token`, as shown in this example URL:

    ```
    https://www.rescuetime.com/api/oauth/data?access_token=ACCESS_TOKEN&by=interval&taxonomy=productivity&interval=hour&restrict_begin=2018-01-01&restrict_end=2018-01-31&format=json&scopes=time_data
    ```


**Oauth2 Access Scopes**

When users connect their accounts to your application via Oauth2, you can specify access scopes to let the user know exactly how you will be using their RescueTime data. Scopes will be requested during the initial authorization request.

The following scopes are currently supported:

-   `time_data:` Access activity history and summary time data or post offline time for the authorized user
-   `category_data:` Access how much time the authorizing user spent in specific categories
-   `productivity_data:` Access how much time the authorizing user spent in different productivity levels
-   `alert_data:` Access the authorizing user's alert history
-   `highlight_data:` Read from and post to the authorizing user's daily highlights list
-   `focustime_data:` Access the authorizing user's Focus Session history and session starting and stopping

## API Endpoints

There are several API endpoints that can be used to access different types of data within RescueTime.

### Analytic Data API:

The Analytic Data API is targeted at bringing developers the prepared and pre-organized data structures already familiar through the reporting views of www.rescuetime.com.

[Detailed documentation for the Analytic Data API](https://www.rescuetime.com/rtx/settings/api/documentation#analytic-api-reference)

### Daily Summary Feed API:

The Daily Summary Feed API provides a structured rollup of daily data about the time logged by a user. This data includes: total time logged, time logged at each productivity level, average productivity pulse, and the time spent in each major category.

[Detailed documentation for the Daily Summary Feed API](https://www.rescuetime.com/rtx/settings/api/documentation#daily-summary-feed-reference)

### Alerts Feed API:

The Alerts Feed API provides a feed of recent occurrences of user-defined alerts. It is a good way to let users define aspects of their data that they care about (“I want to spend more than 5 hours per day on All Productive Time”, for example) and then use that data in an external service.

**Note:** Alerts are a feature of RescueTime premium. This API will return empty results for non-premium users.

[Detailed documentation for the Alerts Feed API](https://www.rescuetime.com/rtx/settings/api/documentation#alerts-feed-reference)

### Highlights Feed API:

The Highlights Feed API provides a feed of recently entered daily highlights.

**Note:** Highlights are a feature of RescueTime premium. This API will return empty results for non-premium users.

[Detailed documentation for the Highlights Feed API](https://www.rescuetime.com/rtx/settings/api/documentation#highlights-feed-reference)

### Highlights POST API:

The Highlights POST API allows users to post a daily highlight entry programatically. An example use case is posting a new highlight whenever a software developer makes a new check-in to their GIT repository.

**Note:** Highlights are a feature of RescueTime premium. This API will return a 400 response for non-premium users.

[Detailed documentation for the Highlights Post API](https://www.rescuetime.com/rtx/settings/api/documentation#highlights-post-reference)

### Focus Session Trigger API:

The Focus Session Trigger API allows users to start Focus Session on their active devices programatically.

**Note:** Focus Session is a feature of RescueTime premium. This API will return a 400 response for non-premium users.

[Detailed documentation for the Focus Session Trigger API](https://www.rescuetime.com/rtx/settings/api/documentation#focustime-trigger-reference)

### Focus Session Feed API:

The Focus Session Feed API provides a feed of recent Focus Session "started" or Focus Session "ended" events. An example use case would be to set a chat application status to "away" when a new Focus Session started event has been detected.

**Note:** Focus Sessions are a feature of RescueTime premium. This API will return a 400 response for non-premium users.

[Detailed documentation for the Focus Session Feed API](https://www.rescuetime.com/rtx/settings/api/documentation#focustime-feed-reference)

### Offline Time POST API:

The Offline Time POST API allows users to post offline time to their account programmatically. An example use case would be to post offline time based on localtion data or meeting times.

**Note:** Offline Time is a feature of RescueTime premium. This API will return a 400 response for non-premium users.

[Detailed documentation for the Offline Time Post API](https://www.rescuetime.com/rtx/settings/api/documentation#offline-time-reference)

## Documentation for the Analytic Data API

RescueTime data is detailed and complicated. The Analytic Data API is targeted at bringing developers the prepared and pre-organized data structures already familiar through the reporting views of www.rescuetime.com. The data is read-only through the webservice, but you can perform any manipulations on the consumer side you want. Keep in mind this is a draft interface, and may change in the future. We do intend to version the interfaces though, so it is likely forward compatible.

The Analytic Data API allows for parameterized access, which means you can change the subject and scope of data, and is especially targeted for developer use. The data can be accessed via the HTTP Query interface in several formats.

The Analytic Data API is presented as a read-only resource accessed by simple GET HTTP queries. This provides maximum flexibility with minimal complexity; you can use the provided API tools, or just research the documented parameters and construct your own queries.

### Service Access

The base URL to reach this HTTP query API is:

-   For connections with an API key: `https://www.rescuetime.com/anapi/data`
-   For Oauth2 connections:
    -   `https://www.rescuetime.com/api/oauth/data` - This is the least restritive url and requires the `time_data` access scope to be granted by the user when the Oauth2 connection is initially set up.
    -   Restricted Report: `https://www.rescuetime.com/api/oauth/overview_data` - This url will restricts the output to top level category data only and requires the `category_data` access scope to be granted by the user when the Oauth2 connection is initially set up.
    -   Restricted Report: `https://www.rescuetime.com/api/oauth/category_data` - This url will restricts the output to sub-category data only and requires the `category_data` access scope to be granted by the user when the Oauth2 connection is initially set up.
    -   Restricted Report: `https://www.rescuetime.com/api/oauth/productivity_data` - This url will restricts the output to productivity data only and requires the `productivity_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

**About restricted reports:** Restricted reports may be useful if your application requires a less-granular rollup of the data, such as the category view, but NOT the actual activities themselves. These restricted reports help ensure the user’s privacy, and may be a preferable option when asking them to link their accounts.

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]
-   `format` - \[ 'csv' | 'json' \]

### Query Parameters

Primary names are chosen for human reading. The short names are for when GET query length is at a premium. The alias is for understanding roughly how it maps into the language used in reporting views, and our own internal nicknames.

|Principle name|Short|Alias|Short|Values|Description|
|---|---|---|---|---|---|
|`perspective`|pv|by|by|\[ 'rank'|'interval' \]|
|`resolution_time`|rs|interval|i|\[ 'month'|'week'|
|`restrict_begin`|rb|date|Sets the start day for data batch, inclusive (always at time 00:00, start hour/minute not supported)
[Format ISO 8601](http://www.w3.org/TR/NOTE-datetime) "YYYY-MM-DD"|
|`restrict_end`|re|date|Sets the end day for data batch, inclusive (always at time 00:00, end hour/minute not supported)
[Format ISO 8601](http://www.w3.org/TR/NOTE-datetime) "YYYY-MM-DD"|
|`restrict_kind`|rk|taxonomy|ty|\[ 'category'|'activity'|
|`restrict_thing`|rt|taxon|tx|name (of category, activity, or overview)|The name of a specific overview, category, application or website. For websites, use the domain component only if it starts with "www", eg. "www.nytimes.com" would be "nytimes.com". The easiest way to see what name you should be using is to retrieve a list that contains the name you want, and inspect it for the exact names.|
|`restrict_thingy`|ry|sub\_taxon|tn|name|Refers to the specific "document" or "activity" we record for the currently active application, if supported. For example, the document name active when using Microsoft Word. Available for most major applications and web sites. Let us know if yours is not.|
|`restrict_source_type`|\[ 'computers'|'mobile'|
|`restrict_schedule_id`|rsi|schedule\_id|s|id (integer id of user's schedule/time filter)|Allows for filtering results by schedule.|

### Output Formats

The Analytic Data API supports CSV and JSON output.

-   `csv` - layout provides rows of comma separated data with a header for column names at top.
-   `json` - returns a JavaScript ready object. It has these properties:
    1.  _notes_ = String, a short explanation of the data envelope
    2.  _row\_headers_ = Array, a label for the contents of each index in a row, in the order they appear in row
    3.  _rows_ = Array X Array, an array of data rows, where each row is itself an array described by the row\_headers

### Example Queries

-   To request information about the user's productivity levels, by hour, for the date of January 1, 2020:

    ```
    https://www.rescuetime.com/anapi/data?key=RESCUE_TIME_API_KEY&perspective=interval&restrict_kind=productivity&interval=hour&restrict_begin=2020-01-01&restrict_end=2020-01-01&format=json
    ```

-   To request a list of time spent in each top level category, ranked by duration, for the date of January 1, 2020:

    ```
    https://www.rescuetime.com/anapi/data?key=RESCUE_TIME_API_KEY&perspective=rank&restrict_kind=overview&restrict_begin=2020-01-01&restrict_end=2020-01-01&format=csv
    ```


## Documentation for the Daily Summary Feed API

The Daily Summary Feed API provides a high level rollup of the the time a user has logged for a full 24 hour period (defined by the user’s selected time zone). This is useful for generating notifications that don’t need to be real-time and don’t require much granularity (for greater precision or more timely alerts, see the Alerts Feed API). This can be used to construct a customized daily progress report delivered via email. The summary can also be used to alert people to specific conditions. _For example, if a user has more than 20% of their time labeled as ‘uncategorized’, that can be used to offer people a message to update their categorizations on the website._

**PLEASE NOTE:** The Daily Summary Feed is ‘point in time‘ data - use the data api or export CSVs from reports if you want your latest categorization results.

### Service Access

The base URL to reach this Daily Summary Feed API is:

-   For connections with an API key: `https://www.rescuetime.com/anapi/daily_summary_feed`
-   For Oauth2 connections: `https://www.rescuetime.com/api/oauth/daily_summary_feed`

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]

### Output Format

The Daily Summary Feed API returns an array of JSON objects representing rollup information for each day logged by the user in the previous two weeks. It does not include the current day, and new summaries for the previous day are available at 12:01 am in the user’s local time zone. The returned value for each daily summary includes the following information:

**General information**

-   `id` - Unix Timestamp representation of the date. Can use useful for deduplication of items in the feed.
-   `date` - Summary Date
-   `productivity_pulse` - Overall productivity pulse. A scale bewteen 0-100

**Percentage representation of time spent in different categories (Float with one decimal place. Min: 0, Max: 100)**

-   `very_productive_percentage`
-   `productive_percentage`
-   `neutral_percentage`
-   `distracting_percentage`
-   `very_distracting_percentage`
-   `all_productive_percentage`
-   `all_distracting_percentage`
-   `uncategorized_percentage`
-   `business_percentage`
-   `communication_and_scheduling_percentage`
-   `social_networking_percentage`
-   `design_and_composition_percentage`
-   `entertainment_percentage`
-   `news_percentage`
-   `software_development_percentage`
-   `reference_and_learning_percentage`
-   `shopping_percentage`
-   `utilities_percentage`

**Numeric representation of HOURS spent in different categories (Float with two decimal places)**

-   `total_hours`
-   `very_productive_hours`
-   `productive_hours`
-   `neutral_hours`
-   `distracting_hours`
-   `very_distracting_hours`
-   `all_productive_hours`
-   `all_distracting_hours`
-   `uncategorized_hours`
-   `business_hours`
-   `communication_and_scheduling_hours`
-   `social_networking_hours`
-   `design_and_composition_hours`
-   `entertainment_hours`
-   `news_hours`
-   `software_development_hours`
-   `reference_and_learning_hours`
-   `shopping_hours`
-   `utilities_hours`

**String representations of time spent in different categories (format - Xh Ym Zs)**

-   `total_duration_formatted`
-   `very_productive_duration_formatted`
-   `productive_duration_formatted`
-   `neutral_duration_formatted`
-   `distracting_duration_formatted`
-   `very_distracting_duration_formatted`
-   `all_productive_duration_formatted`
-   `all_distracting_duration_formatted`
-   `uncategorized_duration_formatted`
-   `business_duration_formatted`
-   `communication_and_scheduling_duration_formatted`
-   `social_networking_duration_formatted`
-   `design_and_composition_duration_formatted`
-   `entertainment_duration_formatted`
-   `news_duration_formatted`
-   `software_development_duration_formatted`
-   `reference_and_learning_duration_formatted`
-   `shopping_duration_formatted`
-   `utilities_duration_formatted`

### Example Queries

-   To request a list of Daily Summaries for a user:

    ```
    https://www.rescuetime.com/anapi/daily_summary_feed?key=RESCUE_TIME_API_KEY
    ```


## Documentation for the Alerts Feed API

The Alerts Feed API is a running log of recently triggered user defined alerts. This is a good event-based representation of data that the user cares about. Alerts are a premium feature and as such the API will _always return zero results for users on the RescueTime Lite plan_.

### Service Access

The base URL to reach the Alerts Feed API is:

-   For connections with an API key: `https://www.rescuetime.com/anapi/alerts_feed`
-   For Oauth2 connections: `https://www.rescuetime.com/api/oauth/alerts_feed` - Requires the `alert_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]
-   `op` - \[ status | list \]

The `op` parameter determines what type of information about the alerts is returned. Passing `op=list` will return a definition list of all the alerts currently defined by the user. This is useful for presenting a list of the user's alerts to allow the user to select an alert\_id value that will be used to filter subsequent queries.

Passing `op` a value of `status` is the default, and will return a list of recently triggered alerts in reverse-chronological order.

### Optional parameters

-   `alert_id` - \[ Integer ID of an alert to filter on \]

### Output Format

When passing `op=list`, The Alerts Feed API will return an array of JSON objects representing the currently active alerts that the user has defined. Each alert object has the following structure:

`{  'id': _integer_ (Unique id of the user-defined alert, can be used for filtering the triggered alerts),  'description': _string_ (The string representation of the alert definition),  'amount': _float_ (the number of hours the user has defined for this alert’s threshold)  }`

When passing `op=status`, the feed returns an array of JSON objects representing the actual triggering of alerts in reverse chronological order. Each object has the following structure:

`{  'id': _integer_ (Unique id for the occurence of the alert triggering),  'alert_id': _integer_ (Unique id of the user-defined alert, can be used to filter alerts),  'description': _string_ (The message generated by the triggered alert),  'created_at': _datetime_ (The time, in user’s selected time zone, that the alert was triggered)  }`

### Example Queries

-   To request a list of for the user's active alerts:

    ```
    https://www.rescuetime.com/anapi/alerts_feed?key=RESCUE_TIME_API_KEY&op=list
    ```

-   To request a list of alert 12345 being triggered for the user:

    ```
    https://www.rescuetime.com/anapi/alerts_feed?key=RESCUE_TIME_API_KEY&op=status&alert_id=12345
    ```


## Documentation for the Highlights Feed API

The Highlights Feed API is a list of recently entered Daily Highlight events. These are user-entered strings that are meant to provide qualitative context for the automatically logged time about the user’s activities. It is often used to keep a log of “what got done today”. Highlights are a premium feature and as such the API will always return zero results for users on the RescueTime Lite plan.

### Service Access

The base URL to reach the Highlights Feed API is:

-   For connections with an API key: `https://www.rescuetime.com/anapi/highlights_feed`
-   For Oauth2 connections: `https://www.rescuetime.com/api/oauth/highlights_feed` - Requires the `highlight_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]

### Output Format

The Highlights Feed API returns an array of JSON objects representing the highlights that have been entered by a user. The returned value for each highlight has the following structure:

`{  'id': _integer_ (Unique id to trigger off of, useful for polling for new items),  'description': _string_ (The text of the highlight),  'date': _date_ (Date that the highlight was entered for, may be different than created_at),  'created_at': _datetime_ (Timestamp when the highlight was entered)  }`

### Example Queries

-   To request a list of for the user's recently entered daily highlights:

    ```
    https://www.rescuetime.com/anapi/highlights_feed?key=RESCUE_TIME_API_KEY
    ```


## Documentation for the Highlights POST API

The Highlights Post API makes it possible to post daily highlight events programmatically as an alternative to entering events manually on RescueTime.com. This is useful for capturing information from other systems and providing a view of the “output” that the user is creating (which is a counterpoint to the “input” attention data that RescueTime logs automatically). Examples include adding highlights whenever a code checkin is done, or marking an item in a to-do list application as complete.

### Service Access

The base URL to reach Highlights POST API is:

-   For connections with an API key: `https://www.rescuetime.com/anapi/highlights_post`
-   For Oauth2 connections: `https://www.rescuetime.com/api/oauth/highlights_post` - Requires the `highlight_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

**Note:** A `POST` request must be used for this API

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]
-   `highlight_date` - A string representing the date the highlight will be posted for. This should be in the format of ‘YYYY-MM-DD’, but a unix timestamp is also acceptable.
-   `description` - A 255 character or shorter string containing the text that will be entered for the highlight. This should be representative of an action that was taken by the user.

### Optional parameters

-   `source` - A **short** string describing the ‘source’ of the action, or the label that should be applied to it. Think of this as a category that can group multiple highlights together in the UI. This is useful when many highlights will be entered. In the reporting UI, they will be collapsed under the expandable source label.

### Output Format

A successful post will return at status code of 200. If there is an error, a status code of 400 will be returned.

### Example Requests

-   To post a highlight about a code checkin the user has just made:

    ```
    POST:
    ```


## Documentation for the FocusTime Trigger API

The FocusTime Trigger API makes it possible to start/end FocusTime on active devices as an alternative to starting/ending it manually from the desktop app. This is useful for automating FocusTime from 3rd party applications. An example would be starting/ending FocusTime at a certain time of day.

**Note:** The RescueTime desktop app syncs with our servers on a 1 minute interval. FocusTime via the API is not a real-time transaction.

### Service Access

The base URL to reach FocusTime Trigger API is:

-   For connections with an API key:
    -   `https://www.rescuetime.com/anapi/start_focustime`
    -   `https://www.rescuetime.com/anapi/end_focustime`
-   For Oauth2 connections:

    -   `https://www.rescuetime.com/api/oauth/start_focustime`
    -   `https://www.rescuetime.com/api/oauth/end_focustime`

    \- Requires the `focustime_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

**Note:** A `POST` request must be used for this API

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]
-   `duration` - An integer representing the length of the FocusTime session in minutes, and must be **a multiple of 5 (5, 10, 15, 20...)**.

    A value of **\-1** can be passed to start FocusTime until the end of the day.

    **Note:** This parameter is not required for the `end_focustime` endpoint.


### Output Format

A successful post will return at status code of 200. If there is an error, a status code of 400 will be returned.

### Example Requests

-   To start FocusTime for 30 minutes on the next desktop app sync:

    ```
    POST:
    ```

-   To end an active FocusTime session:

    ```
    POST:
    ```


## Documentation for the FocusTime Feed API

The FocusTime Feed API is a running log of recently triggered started/ended FocusTime sessions. This is useful for performing 3rd party app interactions whenever a new FocusTime session has started/ended. FocusTime is a premium feature and as such the API will _always return zero results for users on the RescueTime Lite plan_.

### Service Access

The base URL to reach the FocusTime Feed API is:

-   For connections with an API key:
    -   `https://www.rescuetime.com/anapi/focustime_started_feed`
    -   `https://www.rescuetime.com/anapi/focustime_ended_feed`
-   For Oauth2 connections:

    -   `https://www.rescuetime.com/api/oauth/focustime_started_feed`
    -   `https://www.rescuetime.com/api/oauth/focustime_ended_feed`

    \- Requires the `focustime_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

### Required parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]

### Output Format

The FocusTime Feed API returns an array of JSON objects representing the FocusTime started/ended events in reverse chronological order. Each object has the following structure:

`{  'id': _float_ (A UNIX timestamp that represents a unique id for the event),  'duration': _integer_ (The initial length of time in minutes for the FocusTime Session, focustime_started_feed only),  'created_at': _datetime_ (The time, in user’s selected time zone, that the session was started)  }`

### Example Queries

-   To request a list of recent FocusTime started Events:

    ```
    https://www.rescuetime.com/anapi/focustime_started_feed?key=RESCUE_TIME_API_KEY
    ```


## Documentation for the Offline Time POST API

The Offline Time Post API makes it possible to post offline time programmatically as an alternative to entering it manually on RescueTime.com. This is useful for capturing information from other systems. Examples include adding offline time after a meeting on a calendar app, or logging driving time based on location data.

### Service Access

The base URL to reach Offline Time POST API is:

-   For connections with an API key: `https://www.rescuetime.com/anapi/offline_time_post`
-   For Oauth2 connections: `https://www.rescuetime.com/api/oauth/offline_time_post` - Requires the `time_data` access scope to be granted by the user when the Oauth2 connection is initially set up.

**Note:** A `POST` request must be used for this API. Also, offline time posts via the API are limited to a **4 hour** maximum duration, and they can not be created for future dates.

### Required QUERY parameters

-   `key` - [RESCUE_TIME_API_KEY] OR `access_token` - \[ the access token from the Oauth2 Connection \]

### Required JSON keys/values

-   `start_time` - A string representing the date/time the for the start of the offline time block. This should be in the format of ‘YYYY-MM-DD HH:MM:SS’, but a unix timestamp is also acceptable.
-   `end_time/duration` - Either a string representing the date/time the for the end of the offline time block, OR an integer representing the duration of the offline time block in minutes.
-   `activity_name` - A 255 character or shorter string containing the text that will be entered as the name of the activity (e.g. "Meeting", "Driving", "Sleeping", etc).

### Optional JSON keys/values

-   `activity_details` - A 255 character or shorter string containing the text that will be entered as the details of the named activity.

### Output Format

A successful post will return at status code of 200. If there is an error, a status code of 400 will be returned.

### Example Requests

-   To post offline time about a meeting that just ended:

    A `POST` request to the following URL:

    ```
    https://www.rescuetime.com/anapi/offline_time_post?key=RESCUE_TIME_API_KEY
    ```

    with a JSON body of:

    `{  "start_time": "2020-01-01 09:00:00",  "duration": 60,  "activity_name": "Meeting",  "activity_details": "Daily Planning"  }`


## Security Considerations

When you enable access to your data, you generate a key that is saved in RescueTime's system. Essentially, this key provides an alternative read-only authentication to your data. You can keep this key private, in which case it's just as safe as your login and password, or publish it, without exposing your account to other access or risks beyond the scope of the data.

We assume **data api** access is being requested for use on the server side of things, presumably in some code of your own. In this case, you can protect the key as you would any other server side secret. It would never be seen in the browser, but used by your application to retrieve data from us then process and render yourself somehow. So in this case, the key provides read-only access to your data by arbitrary scope (within limits) through parameters.

As a user, **you can revoke access to keys you have created by going to your [Key Management Page](https://www.rescuetime.com/anapi/manage).** Any applications you have connected to via an Oauth2 connection, can be revoked by going to your [connected applications page](https://www.rescuetime.com/oauth/authorized_applications).
