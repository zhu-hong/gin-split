import { ParallelHasher } from 'ts-md5'
import hashWorker from 'ts-md5/dist/md5_worker?url'
import pLimit from 'p-limit'
import axios from 'axios'
import { all } from 'axios'

const input = document.querySelector('input')

input.addEventListener('change', async (e) => {
  /**
   * @type { File }
   */
  const file = e.target.files[0]
  if(!file) return;

  const hasher = new ParallelHasher(hashWorker)

  if(file.size <= 1024*1024*1024) {
    const hash = await hasher.hash(file)
    const fd = new FormData()
    fd.append('hash', hash)
    fd.append('file', file)
    fd.append('fileName', file.name)
    fd.append('frag', 'no')
    axios.post('http://localhost:1122/upload', fd, {
      headers: {
        'Content-Type': file.type,
      },
    })
    return
  }

  const chunkSize = 1024 * 1024 * 1024 / 4 // 分片大小
  const chunkLen = Math.ceil(file.size / chunkSize)
  const chunks = []
  const forHash = []
  
  for (let count = 0; count < chunkLen; count++) {
    const chunk = file.slice(chunkSize*count,chunkSize*(count+1),file.type)
    chunks.push(chunk)
    if(count === 0 || count === chunkLen - 1) {
      forHash.push(chunk)
    } else {
      let center = Math.ceil(chunk.size / 2)
      forHash.push(chunk.slice(0,1024*2,file.type))
      forHash.push(chunk.slice(center-1024,center+1024,file.type))
      forHash.push(chunk.slice(chunk.size-1024*2,chunk.length,file.type))
    }
  }
  console.log(chunks)
  
  let second = 0
  let interval = setInterval(() => {
    second++
    console.log(second)
  }, 1000)
  console.log('start calc hash')
  const hash = await hasher.hash(new Blob(forHash))
  console.log(`hash calc over: ${hash}`)
  clearInterval(interval)
  second=0

  const limit = pLimit(5)
  const reqs = chunks.map((chunk, index) => {
    return limit(() => new Promise((resolve) => {
      const fd = new FormData()
      fd.append('hash', hash)
      fd.append('file', chunk)
      fd.append('fileName', file.name)
      fd.append('index', index.toString())
      fd.append('frag', 'yes')
      axios.post('http://localhost:1122/upload', fd, {
        headers: {
          'Content-Type': file.type,
        },
      }).then(() => {
        resolve(fd)
      })
    }))
  })

  interval = setInterval(() => {
    second++
    console.log(second)
  }, 1000)
  console.log('start upload')
  await Promise.all(reqs)
  console.log(`upload over`)
  clearInterval(interval)
  second=0
})
