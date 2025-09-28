# Changelog

## 1.0.0 (2025-09-28)


### Features

* **diary:** implement CRUD operations for DiaryService ([9c43065](https://github.com/atreya2011/health-management-api/commit/9c430655d39ddfa2adbdba98af0f49afbecdf40c))
* implement body record crud ([a28f5f8](https://github.com/atreya2011/health-management-api/commit/a28f5f89fad85b2d62df32967850b67a26d4ffa1))
* implement columns CRUD and migrate sqlc to pgx/v5 to refactor test helpers ([6727dc5](https://github.com/atreya2011/health-management-api/commit/6727dc5ffaee09c2e3dfdb6b816a9620494b17c9))
* implement CRUD operations for ExerciseRecordService ([bcc6587](https://github.com/atreya2011/health-management-api/commit/bcc65872e7f4f97458b95dbcf2096b0aacfef90d))


### Bug Fixes

* **column:** resolve tag search encoding error and improve test script ([7852c7c](https://github.com/atreya2011/health-management-api/commit/7852c7cac687883af883ef2839d434329e5a4830))
* ensure list and get body records work properly ([e25a8fe](https://github.com/atreya2011/health-management-api/commit/e25a8fe2354ae94fa374b6ec3f71dcd7023422eb))
* ensure that make seed works properly ([5c9a7dd](https://github.com/atreya2011/health-management-api/commit/5c9a7dd0ae9273cbfaccbbe2ad0beab33893a40e))
* update test_body_record_api.sh to dynamically generate JWT token ([cfc3052](https://github.com/atreya2011/health-management-api/commit/cfc305223d70bbdbf5d28edc641ae4ee1dac7484))


### Miscellaneous

* add dependabot for automatic updates ([5a3e568](https://github.com/atreya2011/health-management-api/commit/5a3e568635906bfb5494530438705d50babf25df))
* add server binary to gitignore ([7affcb6](https://github.com/atreya2011/health-management-api/commit/7affcb63e606807bc8805793019218d4fab5135b))
* bootstrap project ([0466929](https://github.com/atreya2011/health-management-api/commit/0466929e53a8e20a57682d6ca737b6d9112095ef))
* **deps:** bump actions/checkout from 3 to 4 ([#2](https://github.com/atreya2011/health-management-api/issues/2)) ([16580ef](https://github.com/atreya2011/health-management-api/commit/16580ef1c2000c445be7f034d17f2d3326c315c1))
* **deps:** bump actions/setup-go from 4 to 5 ([#1](https://github.com/atreya2011/health-management-api/issues/1)) ([64d1fff](https://github.com/atreya2011/health-management-api/commit/64d1fff7c895e9fac85deedef63da248dc2013b0))
* **deps:** bump codecov/codecov-action from 3 to 5 ([#3](https://github.com/atreya2011/health-management-api/issues/3)) ([8825ba3](https://github.com/atreya2011/health-management-api/commit/8825ba358fba955b9860cdb5c2788b0a636d0ce2))
* **deps:** bump golangci/golangci-lint-action from 3 to 7 ([#4](https://github.com/atreya2011/health-management-api/issues/4)) ([ad8c28a](https://github.com/atreya2011/health-management-api/commit/ad8c28a565e495af5e87e04939a5aaa8076924b4))
* go mod tidy ([db1d558](https://github.com/atreya2011/health-management-api/commit/db1d55861a937f9a4d70a449db56009663c38f77))
* go mod tidy ([4c52aa5](https://github.com/atreya2011/health-management-api/commit/4c52aa56a111495eb2043dd863835a8be71511ed))
* remove mocks and use dockertest to test against mock data in a database ([14f6a59](https://github.com/atreya2011/health-management-api/commit/14f6a596e32aa7a80fc9baf26d0dcf436ed55e6b))
* restructure main.go and related files ([f95cbce](https://github.com/atreya2011/health-management-api/commit/f95cbce1752a448f56831a70474cc27273a34939))
* simplify code architecture ([#8](https://github.com/atreya2011/health-management-api/issues/8)) ([be99bee](https://github.com/atreya2011/health-management-api/commit/be99bee912839fbf43856a189249209f31e07d22))
* update buf.yaml linting rules and clean up diary_handler_test.go ([d88698f](https://github.com/atreya2011/health-management-api/commit/d88698f38f542276ed23a26140a320c31ae5731c))
* update CI workflow name to lowercase and streamline Go module handling ([bd23276](https://github.com/atreya2011/health-management-api/commit/bd232765fd8bcb8621c18d14cffe8a4b9352e38f))
* update Makefile and README to reflect new server binary name ([dcf85eb](https://github.com/atreya2011/health-management-api/commit/dcf85eb1da7c22304c5aa8a8a4eb8243e4ab28e3))
* use richgo to colourize tests ([48d8e4f](https://github.com/atreya2011/health-management-api/commit/48d8e4f83815845d2813f52731c6e2efb017f293))


### Documentation

* add database schema diagram to README for better understanding of relationships ([3d77f58](https://github.com/atreya2011/health-management-api/commit/3d77f58a06282740f1766b2b511730f962484ee6))
* expand README with TODO list for features, testing, operations, and security improvements ([d9c313c](https://github.com/atreya2011/health-management-api/commit/d9c313cea0969987a6dfec4a4768a928d5bc452c))
* standardize database schema formatting in README for clarity ([4b2826c](https://github.com/atreya2011/health-management-api/commit/4b2826cae690874e48b14ebcc08b3b68493888c1))
* Update README files to reflect current project state ([d7e341a](https://github.com/atreya2011/health-management-api/commit/d7e341a3f21ad3201042d79d969392ce18531237))
* update README to reflect current project structure ([c7846e0](https://github.com/atreya2011/health-management-api/commit/c7846e01a571e482532d5feb1d9a5f2371f62d71))


### Code Refactoring

* **ci:** consolidate lint, test, and build steps into a single job ([b9cffb7](https://github.com/atreya2011/health-management-api/commit/b9cffb7547c66aa825004167e08c76e9a20cf8fb))
* consolidate test structure and simplify test execution ([241b9d3](https://github.com/atreya2011/health-management-api/commit/241b9d35ce56d48790875b69786cd55e65428e92))
* streamline handler registration in server setup ([0923ec7](https://github.com/atreya2011/health-management-api/commit/0923ec7ca45f57808a0805ba5413fe31aa23d3c9))
* **testutil:** clean up unused imports and relocate test helper functions ([1b32aa2](https://github.com/atreya2011/health-management-api/commit/1b32aa2c4e711ce0f0bc0628f7375beb8c3d0e07))
* update test utilities to use sqlc-generated functions ([f8ad2cc](https://github.com/atreya2011/health-management-api/commit/f8ad2cc16469e949acaf2aa876699e5e9a7082b9))
* use cobra-cli to add serve and seed commands ([d7a96da](https://github.com/atreya2011/health-management-api/commit/d7a96da836cb5e82fe6cc1c558568b9303e23189))
