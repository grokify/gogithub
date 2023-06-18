# GitHub Reports

This project is to auto-generate a GitHub Pages HTML page that produces various reports.

The first report is an open Pull Request list based on the following API:

* API: https://docs.github.com/en/rest/search?apiVersion=2022-11-28
* Example Request: https://api.github.com/search/issues?q=user:grokify%20state:open%20is:pr

It does not appear Pull Request count is separately availble in the repos API call. See more on this API call here:

* Discussion: https://stackoverflow.com/questions/8713596/how-to-retrieve-the-list-of-all-github-repositories-of-a-person