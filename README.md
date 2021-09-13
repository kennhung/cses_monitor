# CSES Monitor

A web crawler for CSES Problem Set.
It will send discord message when someone in the list pass/fail a problem.

## env variables

* `USER_IDS`: id of users, sperated by ','(no space after)
* `FETCH_DELAY`: delay time between fetching each user
* `DISCORD_WEBHOOK`: webhook to send discord message