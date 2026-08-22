#include <curl/curl.h>

#include <iostream>
#include <string>
#include <thread>

#include "threading.h"

static size_t write_callback(
    void* contents,
    size_t size,
    size_t nmemb,
    void* userp
)
{
    const size_t total_size = size * nmemb;

    auto* response =
        static_cast<std::string*>(userp);

    response->append(
        static_cast<char*>(contents),
        total_size
    );

    return total_size;
}

void api_request(std::size_t id)
{
    CURL* curl = curl_easy_init();

    if (!curl)
    {
        std::cerr
            << "Thread "
            << id
            << ": curl init failed\n";

        return;
    }

    std::string response;

    const char* url =
        "http://localhost:5010/api/v1/orders";

    const char* json =
        R"({
          "accountEmail": "string",
          "deliveryAddress": "string",
          "shopItems": [
            {
              "description": "string",
              "id": "string",
              "price": 0,
              "quantity": 0,
              "title": "string"
            }
          ]
        })";

    struct curl_slist* headers = nullptr;

    headers = curl_slist_append(
        headers,
        "Content-Type: application/json"
    );

    headers = curl_slist_append(
        headers,
        "Accept: application/json"
    );

    curl_easy_setopt(
        curl,
        CURLOPT_URL,
        url
    );

    curl_easy_setopt(
        curl,
        CURLOPT_POST,
        1L
    );

    curl_easy_setopt(
        curl,
        CURLOPT_POSTFIELDS,
        json
    );

    curl_easy_setopt(
        curl,
        CURLOPT_HTTPHEADER,
        headers
    );

    curl_easy_setopt(
        curl,
        CURLOPT_WRITEFUNCTION,
        write_callback
    );

    curl_easy_setopt(
        curl,
        CURLOPT_WRITEDATA,
        &response
    );

    curl_easy_setopt(
        curl,
        CURLOPT_TIMEOUT_MS,
        5000L
    );

    curl_easy_setopt(
        curl,
        CURLOPT_NOSIGNAL,
        1L
    );

    const CURLcode result =
        curl_easy_perform(curl);

    if (result == CURLE_OK)
    {
        long status_code = 0;

        curl_easy_getinfo(
            curl,
            CURLINFO_RESPONSE_CODE,
            &status_code
        );


    }
    else
    {
        std::cerr
            << "Thread "
            << id
            << " -> "
            << curl_easy_strerror(result)
            << '\n';
    }

    curl_slist_free_all(headers);

    curl_easy_cleanup(curl);
}

int main()
{
    curl_global_init(
        CURL_GLOBAL_DEFAULT
    );

    const std::size_t task_count =
        10'000;

    perf::Threading threading(
        task_count,
        api_request
    );

    std::cout
        << "Hardware concurrency: "
        << std::thread::hardware_concurrency()
        << '\n';

    std::cout
        << "Tasks: "
        << threading.getTaskCount()
        << '\n';

    std::cout
        << "Workers: "
        << threading.getWorkerCount()
        << '\n';

    threading.go();

    threading.join();

    std::cout
        << "Elapsed: "
        << threading.getElapsedTime()
        << " ms\n";

    curl_global_cleanup();

    return 0;
}