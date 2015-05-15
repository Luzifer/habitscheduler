# Luzifer / habitscheduler

Currently I'm using [HabitRPG](https://habitrpg.com/) for my task management. They are able to manage daily repeating tasks but not tasks recurring using a cron-like scheme or tasks recurring every 6 weeks after their last completion. To handle those tasks this project has been created.

The habitscheduler is a web application providing an API to schedule tasks with cron-like scheme or recurrance expressed in hours after last completion.

## Usage (Docker)

```bash
# docker pull luzifer/habitscheduler
# docker run -ti luzifer/habitscheduler --help
Usage of /go/bin/habitscheduler:
  -habit-token="": API-Token for that HabitRPG user
  -habit-user="": User-ID from API page in HabitRPG
  -listen=":3000": Address incl. port to have the API listen on
  -redis-key="habitrpg-tasks": Key to store the data in
  -redis-url="": Connectionstring to redis server
# docker run -ti luzifer/habitscheduler --redis-url "tcp://auth:...@myhost:6379/0" [...]
```
