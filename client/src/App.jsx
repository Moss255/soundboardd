import { useEffect, useState } from "react";
import Alert from "./Alert";


const App = () => {

  const [files, setFiles] = useState([]);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState('');

  useEffect(() => {
    fetchFiles();
  }, []);

  const handleButtonClick = async (ID) => {

    try {
      const response = await fetch(`http://localhost:8080/play?ID=${ID}`, {
        method: 'GET',
        mode: 'cors',
        headers: {}
      });

      if (response.status > 300) {
        throw 'Unable to GET ' + response.statusText
      }

      const data = await response.blob();

      let blobUrl = URL.createObjectURL(data);

      const audio = new Audio(blobUrl);

      audio.play();
    } catch (error) {
      setMessage('Unable to play requested soundbite')
      setMessageType('is-danger')
    }

  }

  const fetchFiles = async () => {
    const response = await fetch(`http://localhost:8080/files`, {
      method: 'GET',
      mode: 'cors',
      headers: {}
    });

    const data = await response.json();

    setFiles(data);
  }

  return (
    <>
      <Alert message={message} messageType={messageType} clearMessage={() => setMessage(null)} open={message}/>
      <div className='grid'>
        {files.map((file) => {
          return <button className='button' key={file.ID} onClick={() => handleButtonClick(file.ID)} type='button'>{file.Filename}</button>
        })}
      </div>
    </>

  )
}

export default App
