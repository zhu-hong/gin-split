import axios from 'axios'

const input = document.querySelector('input')

input?.addEventListener('change', (e) => {
  /**
   * @type { File }
   */
  const file = e.target.files[0]
  if(!file) return;

  const fd = new FormData()
  fd.append('frag', 'yes')
  fd.append('index', '0')
  fd.append('file', file)

  axios.post('http://localhost:1122/upload', fd)
})
