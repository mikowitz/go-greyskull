# Git merge

Find the current branch name. If it is not `main`, confirm I wish to merge it by asking "Do you want to merge @branch_name into main?" Create a prompt option so I can just hit Return to say yes. If I say yes, checkout main, squash and merge the feature branch in, and delete the feature branch locally and remotely

# Implement prompt

I will pass a prompt number as an argument to this command, and you will carry out the following steps. You should show your progress verbs while doing so, and you do not need to wait to move to the next step once the previous step is completed. NEVER skip any steps!

1. read @greyskull-prompts.md to find the prompt matching the prompt number
2. git checkout a new git branch named after the prompt, all lowercase and words joined by dashes
3. implement the prompt step by step
4. update @greyskull-todo.md to mark all complete steps
5. git add all the changed files
6. create a commit with a descriptive commit message and push it to the git remote origin

# Write tests

I will pass a prompt number as an argument to this command, and you will carry out the following steps. You should show your progress verbs while doing so, and you do not need to wait to move to the next step once the previous step is completed. NEVER skip any steps!

1. read @greyskull-prompts.md to find the prompt matching the prompt number I gave you
2. If and only if you are on the main git branch, check out a new branch named after the prompt, all lowercase and words joined by dashes
3. Implement a comprehensive test suite ONLY based on the test section of the prompt. DO NOT write application code to go along with it. The test suite SHOULD AND MUST FAIL after making these changes
