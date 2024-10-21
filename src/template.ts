export const template = `
Please suggest 10 commit messages, given the following diff:

\`\`\`diff
{{diff}}
\`\`\`

You have to user Semantic Commit Messages way of writing commit messages.
Format: <type>(<scope>): <subject>
- <type> is mandatory and specifies the type of commit.
- <scope> is optional and specifies the place of the commit change.
- <subject> is mandatory and contains succinct description of the change.
- Any line of the commit message cannot be longer 100 characters!
# Types
- feat: A new feature
- fix: A bug fix
- docs: Documentation only changes
- style: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- refactor: A code change that neither fixes a bug nor adds a feature
- perf: A code change that improves performance
- ci: Changes to our CI configuration files and scripts (example scopes: Travis, Circle, BrowserStack, SauceLabs)
- test: Adding missing tests
- chore: anything else
# example
- feat(ng-list): allow custom separator
- fix(ng-list): handle empty arrays
- docs(ng-list): update readme
- style(ng-list): remove whitespace
- refactor(ng-list): use helper function for filter
- perf(ng-list): optimize filter
- ci(ng-list): add Travis CI
- test(ng-list): add missing tests
- chore(ng-list): change linter to eslint
# more example 
also you can write commit like this 
chore: add Oyster build script
docs: explain hat wobble
feat: add beta sequence
fix: remove broken confirmation message
refactor: share logic between 4d3d3d3 and flarhgunnstow
style: convert tabs to spaces
test: ensure Tayne retains clothing

# Notes
you have to be creative and think about the best commit message that can describe the changes in the diff. like if you have a diff that changes the color of a button from red to blue, you can write a commit message like this: style(button): change button color from red to blue. Moreover, you have pick simple words easy to understand.  

Keep in mind you will suggest 10 commit messages. Only 1 will be used. It's better to push yourself (esp to synthesize to a higher level) and maybe wrong about some of the 10 commits because only one needs to be good. I'm looking for your best commit, not the best average commit. It's better to cover more scenarios than include a lot of overlap.

Write your 10 commit messages below, in the format json list format below:
{
  commitMessages: [
    "message",
    "message"
  ]
}
`;
