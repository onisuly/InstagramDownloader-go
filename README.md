# InstagramDownloader-go
Download media from Instagram with go

Usage:
1. Get a valid session id for instagram, login instagram through web browser, get session id from cookie
2. Run instagram downloader
    ```shell
    ./InstagramDownloader-linux-amd64 -username a_instagram_username -sessionId a_valid_session_id
    ```
3. Run with proxy
    ```shell
    ./InstagramDownloader-linux-amd64 -username a_instagram_username -sessionId a_valid_session_id -proxy socks5://127.0.0.1:1080
    ```
 