# Contribution Guidelines

## Table of Contents

- [Contribution Guidelines](#contribution-guidelines)
  - [Introduction](#introduction)
  - [Bug reports](#bug-reports)
  - [Discuss your design](#discuss-your-design)
  - [Testing redux](#testing-redux)
  - [Vendoring](#vendoring)
  - [Building Ledisdb](#building-ledisdb)
  - [Code review](#code-review)
  - [Developer Certificate of Origin (DCO)](#developer-certificate-of-origin-dco)
  - [Copyright](#copyright)

## Introduction

This document explains how to contribute changes to the Ledisdb project.

## Bug reports

Please search the issues on the issue tracker with a variety of keywords
to ensure your bug is not already reported.

If unique, [open an issue](https://github.com/siddontang/ledisdb/issues/new)
and answer the questions so we can understand and reproduce the
problematic behavior.

To show us that the issue you are having is in ledisdb itself, please
write clear, concise instructions so we can reproduce the behavior—
even if it seems obvious. The more detailed and specific you are,
the faster we can fix the issue. Check out [How to Report Bugs
Effectively](http://www.chiark.greenend.org.uk/~sgtatham/bugs.html).

Please be kind, remember that Gitea comes at no cost to you, and you're
getting free help.

## Discuss your design

The project welcomes submissions. If you want to change or add something,
please let everyone know what you're working on—[file an issue](https://github.com/go-gitea/gitea/issues/new)!
Significant changes must go through the change proposal process
before they can be accepted. To create a proposal, file an issue with
your proposed changes documented, and make sure to note in the title
of the issue that it is a proposal.

This process gives everyone a chance to validate the design, helps
prevent duplication of effort, and ensures that the idea fits inside
the goals for the project and tools. It also checks that the design is
sound before code is written; the code review tool is not the place for
high-level discussions.

## Testing redux

Before submitting a pull request, run all the tests for the whole tree
to make sure your changes don't cause regression elsewhere.

Here's how to run the test suite:

- `make vet`
- `make test`

## Vendoring

We keep a cached copy of dependencies within the `vendor/` directory,
managing updates via [Modules](https://golang.org/cmd/go/#hdr-Module_maintenance).

Pull requests should only include `vendor/` updates if they are part of
the same change, be it a bugfix or a feature addition.

The `vendor/` update needs to be justified as part of the PR description,
and must be verified by the reviewers and/or merger to always reference
an existing upstream commit.

You can find more information on how to get started with it on the [Modules Wiki](https://github.com/golang/go/wiki/Modules).

## Building Ledisdb

```
make build
```

## Code review

Changes to Ledisdb must be reviewed before they are accepted—no matter who
makes the change, even if they are an owner or a maintainer. We use GitHub's
pull request workflow to do that.

Please try to make your pull request easy to review for us. And, please read
the *[How to get faster PR reviews](https://github.com/kubernetes/community/blob/261cb0fd089b64002c91e8eddceebf032462ccd6/contributors/guide/pull-requests.md#best-practices-for-faster-reviews)* guide;
it has lots of useful tips for any project you may want to contribute.
Some of the key points:

* Make small pull requests. The smaller, the faster to review and the
  more likely it will be merged soon.
* Don't make changes unrelated to your PR. Maybe there are typos on
  some comments, maybe refactoring would be welcome on a function... but
  if that is not related to your PR, please make *another* PR for that.
* Split big pull requests into multiple small ones. An incremental change
  will be faster to review than a huge PR.

## Developer Certificate of Origin (DCO)

We consider the act of contributing to the code by submitting a Pull
Request as the "Sign off" or agreement to the certifications and terms
of the [DCO](DCO) and [MIT license](LICENSE). No further action is required.
Additionally you could add a line at the end of your commit message.

```
Signed-off-by: Joe Smith <joe.smith@email.com>
```

If you set your `user.name` and `user.email` git configs, you can add the
line to the end of your commit automatically with `git commit -s`.

We assume in good faith that the information you provide is legally binding.

## Copyright

Code that you contribute should use the standard copyright header:

```
// Copyright 2020 The Ledisdb Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
```

Files in the repository contain copyright from the year they are added
to the year they are last changed. If the copyright author is changed,
just paste the header below the old one.
