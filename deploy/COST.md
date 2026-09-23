# Cost Scenario Analysis (Task 5)

> This is where you answer Task 5. Up to 1,000 words. If you can get real data, give numbers and the queries you would run. If you cannot, write down what you are assuming, how you would check it, and what would change if the assumption turned out to be wrong. Tie each conclusion to a specific piece of data and to how you would check it — a general list of cost-saving tips is not what we are looking for.

## 1. Where You Start

What do you look at first to decide whether +30% is even surprising? Which unit do you measure in — cost per event, per request, per active user — and why that one? How do you tell 
"we are simply doing more business" apart from "we are wasting money"?

>Likely from the total amount of events, since we already know that client requests mechanism was changed. If we still have about 5% increase of events in DynamoDB then
>cost increase connected to other layers, if not - maybe it's just expected behavior, because increase amount of events to 30% will also increse the cost for other services too, 
>not just DynamoDB. Most interesting to compare is cost per request before and after the release. So i would start investigating request-event relationship, then compare that with 
>the actual DynamoDB, S3, EKS and ELB usage. Increasin all services cost separately on about 30% could be normal and expected.


## 2. Where the Money Went

Walk through one chain of *what you saw → what you think it is → the exact data that would prove it*, until the 30% is broken into named pieces. Say which data you would use at each step:

- which dimension you split by first — service, usage type, account, region, tag;

> Probably service first, because interesting how cost distributed between them after release, then take the most suspicious and try usage type on it. Then do the same for the other servicesas well.

- the 40% of resources with no tags: how do you work out who they belong to?

> I would use recource ID from the CUR and try to find out where it belongs to. Also would try IaC, where the resource could have an ownership, not sure which IaC tool was used though. Cloudtrail was not mentioned in available tools, but it could help too.

- four services rising at once: one shared cause, or several separate ones, and how would you tell?

>if the usage increase started immediately after deploy on all 4, then it was the exact reason, if not - should investigate separately each case.

- the other explanations you have to rule out before you trust your own: the different number of days in the two months, how those Reserved Instance and Savings Plan payments land in the 
bill, one-off charges.

>Surely I would look for charges that don't represent recurring workload, in particular those RI and SP, so make sure that this 30% increase is actually from increased workload.

## 3. What You Would Do

Pick one of the three situations in the README. Give a first step you could start on Monday, a fallback if it does not work, and what you give up by choosing it.

Then separate the one-off fix from the thing that keeps the cost from creeping back up, and say how you would keep it from creeping back.

>I would start with DynamoDB and if it turns out that the reason is actually increased number of writes, caused consequent increase to other services. If so, there's a reason to talk with development team
>and discuss the possible hotfix or bring it to next release. If this plan does not work, optimizations to 15% seems to be still a way to go, versus buying resources for the table which is moving 
>away.

## 4. Something You Have Done Before (Optional)

One cost cut you actually made: the numbers before and after, and how you convinced yourself the saving came from your change rather than from business volume moving on its own.
>Many of projects like above one, once I remember even replacing Jfrog artifactory with ECR and CodeArtifact, which was not just resource saving, but also a license.
