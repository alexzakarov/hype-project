#include "timer.h"

namespace perf
{
    Timer::Timer()
    {
        start();
    }

    Timer::~Timer()
    {
        stop();
    }

    void Timer::start()
    {
        startTimePoint =
            std::chrono::high_resolution_clock::now();
    }

    void Timer::stop()
    {
        endTimePoint =
            std::chrono::high_resolution_clock::now();
    }

    double Timer::getElapsedTime()
    {
        const auto elapsed =
            std::chrono::duration_cast<std::chrono::microseconds>(
                endTimePoint - startTimePoint
            );

        elapsed_milli_seconds =
            elapsed.count() / 1000.0;

        return elapsed_milli_seconds;
    }
}