# Distributed Search Engine (Google Search Clone)

### Feature Wishlist
- Rendering dynamic webpages - Current version only loads regular html webpage, and doesn't run JavaScript. My plan is to use a library to add support for dynamic websites which use JavaScipt to load content.
- URL tracking using URL shortener - All the search results would have a proxy URL which redirects to the actual result. So each time someone clicks something, we know what they searched and what they clicked on. This won't be any useful for the small number of users this project will have. But that's not the point of building it, I'm doing it for the satisfying my curiosity.
- Autocomplete - I have no idea how to do this, but I know it's possible. I can imagine one service generating all the searches and the other generating a trie out of those searches. Another serivce can retrieve data from the trie and interface with the frontend. Simple REST API won't cut it here, and gRPC isn't supported by frontend, so I'll have to think of something else. This might reuiqre refactoring the frontend service to use ReactJS instead of just plain Go.
- Monitoring and logging - I want to have a Grafana dahsboard with all the relevant metrics in one place. 
- Infrastructure As Code support - I'm planning to split each service into it's only repo, and have some Terrraform code to deploy all that on GCP/AWS.
