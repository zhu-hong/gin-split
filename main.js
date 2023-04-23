import { ParallelHasher } from 'ts-md5'
import hashWorker from 'ts-md5/dist/md5_worker?url'
import pLimit from 'p-limit'
import axios from 'axios'

const input = document.querySelector('input')

const hasher = new ParallelHasher(hashWorker)
const chunkSize = 1024 * 1024 * 1024 / 4 // 分片大小

input.addEventListener('change', async (e) => {
  /**
   * @type { File }
   */
  const file = e.target.files[0]
  if (!file) return;

  if (file.size <= chunkSize) {
    let second = 0
    let interval = setInterval(() => {
      second++
      console.log(second)
    }, 1000)
    console.log('hash start')
    const hash = await hasher.hash(file)
    console.log(`hash over: ${hash}`)
    second = 0
    clearInterval(interval)

    const checkRes = await axios.get('http://localhost:1122/CheckFile', {
      params: {
        hash,
        fileName: file.name,
      },
    })

    if (checkRes.data.exist === 1) {
      console.log(`file exist: http://127.0.0.1:1122/File/${checkRes.data.file}`)
      return
    }

    const fd = new FormData()
    fd.append('file', file)
    fd.append('hash', hash)
    fd.append('fileName', file.name)
    fd.append('frag', 'no')

    interval = setInterval(() => {
      second++
      console.log(second)
    }, 1000)
    console.log('start upload')
    const res = await axios.post('http://localhost:1122/File', fd, {
      headers: {
        'Content-Type': file.type,
      },
    })
    second = 0
    clearInterval(interval)
    console.log(`upload over: http://127.0.0.1:1122/File/${res.data.file}`)
    return
  }

  const chunkLen = Math.ceil(file.size / chunkSize)
  const chunks = []
  const forHash = []

  for (let count = 0; count < chunkLen; count++) {
    const chunk = file.slice(chunkSize * count, chunkSize * (count + 1), file.type)
    chunks.push(chunk)
    if (count === 0 || count === chunkLen - 1) {
      forHash.push(chunk)
    } else {
      let center = Math.ceil(chunk.size / 2)
      forHash.push(chunk.slice(0, 1024 * 2, file.type))
      forHash.push(chunk.slice(center - 1024, center + 1024, file.type))
      forHash.push(chunk.slice(chunk.size - 1024 * 2, chunk.length, file.type))
    }
  }

  let second = 0
  let interval = setInterval(() => {
    second++
    console.log(second)
  }, 1000)
  console.log('hash start')
  const hash = await hasher.hash(new Blob(forHash))
  console.log(`hash over: ${hash}`)
  clearInterval(interval)
  second = 0

  const checkRes = await axios.get('http://localhost:1122/CheckFile', {
    params: {
      hash,
      fileName: file.name,
    },
  })

  if (checkRes.data.exist === 1) {
    console.log(`file exist: http://127.0.0.1:1122/File/${checkRes.data.file}`)
    return
  }

  const limit = pLimit(5)
  const reqs = chunks.map((chunk, index) => {
    if (checkRes.data.chunks.includes(index)) return Promise.resolve('')

    return limit(() => new Promise((resolve, reject) => {
      const fd = new FormData()
      fd.append('file', chunk)
      fd.append('hash', hash)
      fd.append('fileName', file.name)
      fd.append('index', `${index}`)
      fd.append('frag', 'yes')
      axios.post('http://localhost:1122/File', fd, {
        headers: {
          'Content-Type': file.type,
        },
      }).then(() => {
        resolve(fd)
      }).catch((rea) => {
        reject(rea)
      })
    }))
  })

  interval = setInterval(() => {
    second++
    console.log(second)
  }, 1000)
  console.log('upload start')
  await Promise.all(reqs)
  console.log(`upload over`)
  clearInterval(interval)
  second = 0

  interval = setInterval(() => {
    second++
    console.log(second)
  }, 1000)
  console.log('merge start')
  const res = await axios.post('http://localhost:1122/MergeFile', {
    hash,
    fileName: file.name,
  })
  console.log(`merge over: http://127.0.0.1:1122/File/${res.data.file}`)
  second = 0
  clearInterval(interval)
})
