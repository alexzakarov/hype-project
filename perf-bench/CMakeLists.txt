cmake_minimum_required(VERSION 4.0)
project(perf)

set(CMAKE_CXX_STANDARD 14)

find_package(cpr CONFIG REQUIRED)
find_package(benchmark CONFIG REQUIRED)
include_directories(.)

add_executable(perf
        main.cpp
        timer.cpp
        timer.h
        threading.cpp
        threading.h)

target_link_libraries(perf PRIVATE cpr::cpr cpr::cpr)
target_link_libraries(perf PRIVATE benchmark::benchmark benchmark::benchmark_main)