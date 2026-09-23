# Implementation Notes and Decision Record

> Record only assumptions, decisions, and evidence from this submission. Reference specific files, jobs, commands, or runtime results. Keep the document concise and aim for no more than 1,000 words.

## 1. Key Assumptions

1. Provided Code Quality & Scope: Assumed the provided application code was tested and functional. If defects were discovered, extra time would be required to patch or adapt alternative implementations. Because this is a demonstration task, exposing metrics rather than implementing full operational logic is sufficient. Minimal custom coding was performed, focusing directly on the specific project scope within the estimated timeframe.
2. Local Deployment Boundary: Deployment runs strictly in a local environment, avoiding immediate cloud security risks and credential management overhead. However, a stub/skeleton cloud workflow was created to decouple future cloud CI/CD pipelines from local runner executions.
3. CI/CD Authentication & Security: Assumed no long-lived static tokens are required, as GitHub Actions operates within the local runner context. If authentication is required later, standard GitHub Secrets/Variables will be used per environment. Security tools (such as SonarQube or JFrog Xray) were intentionally omitted to keep the pipeline lightweight.

---
#List three key assumptions that your implementation depends on. These may concern the deployment boundary, team workflow, traffic patterns, or external platform capabilities.
#For each assumption, explain why it was needed and what would need to change if it proved false. Do not present assumptions as known facts.

## 2. Delivery Path

The full delivery pipeline and architectural details are documented in `github-ci-cd-description.md`.

Executing local deployments requires an active GitHub Actions self-hosted runner.

###Please not that non-standard host ports were used to avoid collisions with pre-existing services in my environment:

task-api        0.0.0.0:8088->8080/tcp, [::]:8088->8080/tcp
prometheus      0.0.0.0:9091->9090/tcp, [::]:9091->9090/tcp
grafana 0.0.0.0:3000->3000/tcp, [::]:3000->3000/tcp
loki    0.0.0.0:3100->3100/tcp, [::]:3100->3100/tcp

#Starting with a pull or merge request, describe the jobs the code passes through, the event that publishes the image, how the artifact is identified, where it is deployed, and the smallest unit that can be rolled back.

#Reference the relevant workflow jobs, deployment commands, and image identifiers. If a step could not be run because an external environment was unavailable, state the validation boundary clearly.

## 3. One Actual Validation or Investigation

### Image Size Optimization

The primary technical constraint was reducing the application container image size under 15 MiB.

The default build produced an image of approximately 90 MiB, exceeding the target footprint.
Standard base images contained unnecessary utilities and OS binaries that increased the build size.
Even using smallest image, there was no certainity that application and OS will fit into 15 MiB, so had to make some experimenting on that.
Switching to a build using `alpine scratch` base image did the trick. Also the Go binary was compiled with symbol table strip flags:
  ```bash
  CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/task-api .
  ```
The optimized final build ran successfully on `scratch` well below the 11 MiB threshold.

Another risky part was 2-3 hours to complete, which was a bit tight timeline in my conditions.

## 4. Two Engineering Trade-offs

###Metric Granularity (HTTP Method Splitting) sound like a good idea for this project. 
But tight overall timeline (estimated 2–3 hours) and the fact that it did not work right away and was not strictly requested,
the choice between
  1. Refactor metrics middleware to track HTTP request details by method (`GET`, `POST`, `PUT`).
  2. Implement basic aggregated request counts.

For now  method-level breakdown for deferred to stay within time constraints and prioritized basic metric generation.
Will result in reduced visibility into request method distribution during traffic analysis.
Highly recommended for troubleshooting requirements in production.

###Grafana Dashboard Scope & Alerting
Fast turnaround required for demonstration output.
Configure comprehensive Grafana dashboards with automated alert routing would be quite nice,
because now we have very limited dashboard setup to basic visualizations required for verification.
But, since minimal functional dashboards covering core metrics necessary for evaluation and no alerting is required,
this idea was also currently abandoned. I did install `loki` container with raw logs provided to the on-call engineer.
For sure absence of proactive automated alerts for error spikes, or latency degradations needs to be fixed in transitioning 
to production deployment with SLA/SLO monitoring.

---

## 5. Actual Time Spent

#The suggested effort is 2–3 hours, not a hard limit.

- Actual time spent: 
     4–5 hours (distributed across 2 days after business hours). On day 1 I did the basic deployment and monitoring part, on the second day to improve the main.go and finished improving pipeline.
- Work deliberately left out, and why:
     Comprehensive user documentation was excluded to save time. Primary pipeline descriptions and schemas were consolidated into `github-ci-cd-description.md`.
- What you would do next with another 60 minutes:
  1. Add method labels (`POST`, `PUT`, `GET`) to application metrics.
  2. Implement standard Grafana alert thresholds for HTTP errors.

## 6. Use of AI

#If you used AI:

- listed transcript files under `deploy/ai-transcripts/`;
#- identify the tool and model for each session when known;
#- describe one specific output that you changed or rejected and the evidence that helped you find the problem.

#The transcript files must contain every prompt and visible response, as required by the repository README. If you did not use AI, write “Not used.”
