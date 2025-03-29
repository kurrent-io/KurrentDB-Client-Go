# Contributing to KurrentDB Client Go

## Working with the Git

The client uses `main` as the main development branch. It contains all changes to the upcoming release. 

We do our best to ensure a clean history. To do so, we ask contributors to squash their commits into a set or single
logical commit.

**To contribute to KurrentDB Client Go**:

1. Fork the repository.
2. Create a feature branch from the `main` branch.
3. It's recommended that feature branches use a rebase strategy (see more in [Git documentation](https://git-scm.com/book/en/v2/Git-Branching-Rebasing)). We also highly recommend using clear commit messages that represent the unit of change.
4. Rebase the latest source branch from the main repository before sending PR.
5. When ready to create the Pull Request on GitHub [check to see what has previously changed](https://github.com/kurrent-io/KurrentDB-Client-Go/compare).

## Samples

Code samples are in the [`samples`](/samples) folder. They're orchestrated in the separate [documentation repository](https://github.com/kurrent-io/documentation). The Kurrent Documentation site is publicly accessible at https://docs.kurrent.io/.

## Code style

1. **Formatting**: Use `gofmt` to format your code. This tool is the standard for Go code formatting and ensures consistency across the codebase.
2. **Linting**: Use `golint` to check for style mistakes. This tool helps maintain code quality by enforcing Go idioms and best practices.
3. **Imports**: Organize imports into three groups: standard library packages, third-party packages, and local packages. Separate these groups with a blank line.
4. **Comments**: Use comments to explain the purpose of the code. Exported functions and types should have comments that start with the name of the function or type.
5. **Naming**: Follow Go naming conventions. Use camelCase for variable names, PascalCase for exported names, and ALL_CAPS for constants.
6. **Error Handling**: Handle errors explicitly. Check for errors and handle them appropriately. Do not ignore errors.
7. **Testing**: Write tests for your code. Use the `testing` package and aim for high test coverage. Ensure tests are clear and concise.
8. **Concurrency**: Use goroutines and channels appropriately. Avoid common pitfalls

## Licensing and legal rights

By contributing to KurrentDB Client Go:

1. You assert that contribution is your original work
2. You assert that you have the right to assign the copyright for the work
3. You accept the [Contributor License Agreement](https://gist.github.com/eventstore-bot/7a1e56c21e81f44a625a7462403298bf) (CLA) for your contribution
4. You accept the [License](LICENSE.md)
