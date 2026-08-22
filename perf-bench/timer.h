//
// Created by Muhammed on 8/19/2026.
//
#ifndef PERF_TIMER_H
#define PERF_TIMER_H

#include <chrono>

namespace perf
{
    class Timer
    {
    public:
        Timer();
        ~Timer();

        void start();
        void stop();
        double getElapsedTime();

    private:
        std::chrono::high_resolution_clock::time_point startTimePoint;
        std::chrono::high_resolution_clock::time_point endTimePoint;

        long long elapsed_milli_seconds = 0;
    };
}

#endif