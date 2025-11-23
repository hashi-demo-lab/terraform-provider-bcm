## User Input

```text
$ARGUMENTS
```

## Outline

work on $ARGUMENT autonomously through the speckit stages, first confirm the issue is valid, when you start mark the issue to in-progress using the label in-progress, update the github issue with comments when you start and finish each
speckit stage with a short summary, solve all questions independently. once fully tested ensure full test coverage and generate terraform examples and validate, when terraform test completed successfully then generate the docs using the
script and create a PR once all successful otherwise work to resolve issues

ToDo List - autonomously

1. `/speckit.specify` - Create feature specification from natural language
2. commit and update Git issue
3. `/speckit.clarify` - clarification about underspecified areas - refine specifications
4. `/speckit.plan` - Generate implementation plan with design artifacts
5. commit and update Git issue
6. `/speckit.tasks` - Generate actionable task list
7. commit and update Git issue
8. `/speckit.analyze` - Analyze spec for TDD compliance
9. commit and update Git issue
10. `/speckit.implement` - Execute all tasks to implement the feature, use tdd, make sure you use the terraform-provider-design skill, resolve issues independently, validate all tests passing, validate all examples passing, documentation generated. validate linking and fix any linting issues Create PR with summary
