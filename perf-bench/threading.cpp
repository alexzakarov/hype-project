#include "threading.h"

#include <algorithm>
#include <utility>

namespace perf
{
    Threading::Threading(
        std::size_t task_count,
        Task task,
        std::size_t requested_worker_count
    )
        : task_count(task_count),
          worker_count(requested_worker_count),
          task(std::move(task)),
          timer()
    {
        if (worker_count == 0)
        {
            worker_count = std::thread::hardware_concurrency();
        }

        if (worker_count == 0)
        {
            worker_count = 1;
        }

        if (task_count == 0)
        {
            worker_count = 0;
        }
        else
        {
            worker_count =
                std::min(worker_count, task_count);
        }
    }

    Threading::~Threading()
    {
        join();
    }

    void Threading::go()
    {
        if (task_count == 0)
        {
            return;
        }

        next_task = 0;
        completed_tasks = 0;
        stop = false;
        running = true;

        timer.start();

        for (std::size_t i = 0; i < worker_count; ++i)
        {
            workers.emplace_back(
                &Threading::worker,
                this
            );
        }

        cv.notify_all();
    }

    void Threading::worker()
    {
        while (true)
        {
            std::size_t task_id;

            {
                std::unique_lock<std::mutex> lock(mutex);

                cv.wait(
                    lock,
                    [this]()
                    {
                        return stop ||
                               next_task < task_count;
                    }
                );

                if (stop)
                {
                    return;
                }

                task_id = next_task++;
            }

            task(task_id);

            {
                std::lock_guard<std::mutex> lock(mutex);

                ++completed_tasks;

                if (completed_tasks == task_count)
                {
                    stop = true;

                    cv.notify_all();
                }
            }
        }
    }

    void Threading::join()
    {
        for (auto& worker : workers)
        {
            if (worker.joinable())
            {
                worker.join();
            }
        }

        workers.clear();

        if (running)
        {
            timer.stop();
            running = false;
        }
    }

    double Threading::getElapsedTime()
    {
        return timer.getElapsedTime();
    }

    std::size_t Threading::getTaskCount() const
    {
        return task_count;
    }

    std::size_t Threading::getWorkerCount() const
    {
        return worker_count;
    }
}