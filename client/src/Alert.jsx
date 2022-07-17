const Alert = (props) => {

    const notification = props.open ? `notification ${props.messageType} alert-visible` : 'alert-hidden'

    return (
        <div className={notification}>
            <button className="delete" onClick={() => props.clearMessage()}></button>
            <span>${props.message}</span>
        </div>
    )
}

export default Alert;
