api=workqueue

# /queue/0,1 <-- 2 workers
# task.1,task.3,task.5 <-- 3 tasks per iter

expected=("Processing 6 task.1.txt" "Task completed task.1.txt" "Task completed task.1.txt with return code 0" "Task completed task.3.txt" "Task completed task.3.txt with return code 64")
expectTaskFailure=64

NUM_DESIRED_OUTPUTS=3