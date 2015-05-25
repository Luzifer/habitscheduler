# Luzifer / habitscheduler

[![API Documentation](http://badge.luzifer.io/v1/badge?title=API&text=Documentation)](http://ipfs.hub.luzifer.io/ipns/swagger.luzifer.io/?url=Qmeq7Qn7S8Re2rbHeWB37zaeP3oys3paZHEjSH9ctk5P8E)

Currently I'm using [HabitRPG](https://habitrpg.com/) for my task management. They are able to manage daily repeating tasks but not tasks recurring using a cron-like scheme or tasks recurring every 6 weeks after their last completion. To handle those tasks this project has been created.

The habitscheduler is a web application providing an API to schedule tasks with cron-like scheme or recurrance expressed in hours after last completion.

## Usage (Docker)

```bash
# docker pull luzifer/habitscheduler
# docker run -ti luzifer/habitscheduler --help
Usage of /go/bin/habitscheduler:
  -cron-create="0 * * * * *": Cron entry for creating new tasks
  -cron-persist="0 * * * * *": Cron entry for saving data to Redis
  -cron-update="10 */5 * * * *": Cron entry for fetchin task updates from HabitRPG
  -habit-token="": API-Token for that HabitRPG user
  -habit-user="": User-ID from API page in HabitRPG
  -listen=":3000": Address incl. port to have the API listen on
  -redis-key="habitrpg-tasks": Key to store the data in
  -redis-url="": Connectionstring to redis server
# docker run -ti luzifer/habitscheduler --redis-url "tcp://auth:...@myhost:6379/0" [...]
```
