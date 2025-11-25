## User Input

```text
$ARGUMENTS
```

## Outline

work on $ARGUMENT autonomously

Workflow - autonomously complete the tasks, follow the steps in each speckit workflow.
speckit will work from a new branch, the instructions are in the specify step.

0. first confirm the gh issue is valid, when you start mark the issue to in-progress using the label in-progress, update the github issue with comments when you start and finish each speckit stage with a short summary
1. `/speckit.specify` - Create feature specification from natural language and continue to next stage
2. commit and update Git issue
3. `/speckit.plan` - Generate implementation plan with design artifacts
4. commit and update Git issue and continue to next stage
5. `/speckit.tasks` - Generate actionable task list
6. commit and update Git issue and continue to next stage
7. `/speckit.analyze` - Analyze spec for TDD compliance
8. commit and update Git issue and continue to next stage
9. `/speckit.implement` - Execute all tasks to implement the feature, use tdd, make sure you use the terraform-provider-test skill to check test gaps, resolve issues independently, validate full tests passing, validate all examples passing, documentation generated. use the skill to check test gaps and resolve, validate linting and fix any linting issues

once fully tested ensure full test coverage and generate terraform examples and validate, when terraform test completed successfully then generate the docs using the
script and create PR with summary once all successful otherwise work to resolve issues
