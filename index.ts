import { cheerio } from "https://deno.land/x/cheerio@1.0.2/mod.ts";
import "https://deno.land/x/dotenv/load.ts";

enum ProblemStatus {
  None,
  Full,
  Zero,
}

type ProblemInfo = {
  status: ProblemStatus;
  name: string;
};
type UserStatus = Record<string, ProblemInfo>;

const nameMap: Record<string, string> = {};

async function getUserInfo(userId: string): Promise<UserStatus> {
  const req = await fetch(`https://cses.fi/problemset/user/${userId}/`);
  if (req.ok) {
    const html = await req.text();

    const $ = cheerio.load(html);

    const result: UserStatus = {};

    $(".task-score").each((i, e) => {
      const href = $(e).attr("href");

      const title = $(e).attr("title");

      if (href) {
        result[href] = {
          status: $(e).hasClass("full")
            ? ProblemStatus.Full
            : $(e).hasClass("zero")
            ? ProblemStatus.Zero
            : ProblemStatus.None,
          name: title || "",
        };
      }
    });

    const name = $("h2").text().split(" ")[2];

    nameMap[userId] = name || "unknown";

    return result;
  }

  return {};
}

const userList = (Deno.env.get("USER_IDS")?.split(",")) || [];
const fetchDelay = isNaN(parseInt(Deno.env.get("FETCH_DELAY") || ""))
  ? 2000
  : parseInt(Deno.env.get("FETCH_DELAY") || "");

const WEBHOOK = Deno.env.get("DISCORD_WEBHOOK");

const store: Record<string, string> = {};

function sendNotification(
  userName: string,
  probName: string,
  probURL: string,
  status: ProblemStatus,
) {
  if (WEBHOOK) {
    fetch(WEBHOOK, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        content: `${
          status === ProblemStatus.Full ? ":white_check_mark:" : ":x:"
        } \`${userName}\` ${
          status === ProblemStatus.Full ? "pass" : "fail on"
        } ||**${probName}**||\nhttps://cses.fi${probURL}`,
      }),
    });
  }
}

function run(index: number) {
  if (index < userList.length) {
    const user = userList[index];

    getUserInfo(user).then((req) => {
      if (store[user]) {
        const prev: UserStatus = JSON.parse(store[user]);

        Object.keys(req).forEach((key) => {
          if (req[key].status !== prev[key].status) {
            sendNotification(
              nameMap[user],
              req[key].name,
              key,
              req[key].status,
            );
          }
        });
      }

      store[user] = JSON.stringify(req);
      setTimeout(() => {
        run(index + 1);
      }, fetchDelay);
    });
  } else {
    run(0);
  }
}

if (userList.length > 0) {
  run(0);
}
