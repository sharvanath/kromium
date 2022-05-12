# Checkpointing and parallel workers
* Every worker starts with a random UUID. Kromium assumes that the transform description hash uniquely identifies the change (this will always hold true as long as the logic in the transforms does not change, to handle that we can simply delete the objects in the checkpoint directory). Each worker writes one file after it has finished processing, named <transformhash_UUID>.
* Each worker picks a random UUID when it starts. When a worker starts it picks a set of X random objects to work on. If it notices the files have already been worked on, it finds a different set. If each set size is small compared to the total no. of files, the hope is that duplicate work will be minimal. Each worker also tries to compact the existing bitmaps by writing it in its own state file and deleting the older ones it subsumes.