#ifndef PERF_THREADING_H
#define PERF_THREADING_H

#include <condition_variable>
#include <cstddef>
#include <functional>
#include <thread>
#include <vector>

#include "timer.h"

namespace perf
{
    class Threading
    {
    public:

        using Task = std::function<void(std::size_t)>;

        Threading(
            std::size_t task_count,
            Task task,
            std::size_t worker_count = 0
        );

        ~Threading();

        void go();
        void join();

        double getElapsedTime();

        std::size_t getTaskCount() const;
        std::size_t getWorkerCount() const;

    private:

        void worker();

        std::size_t task_count;
        std::size_t worker_count;

        Task task;

        std::vector<std::thread> workers;

        std::size_t next_task = 0;
        std::size_t completed_tasks = 0;

        std::mutex mutex;
        std::condition_variable cv;

        bool stop = false;
        bool running = false;

        Timer timer;
    };
}

#endif