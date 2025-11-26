import React from 'react'

export default function ItineraryDay({ day }){
  return (
    <div>
      {day && day.length>0 ? (
        <ul>
          {day.map((place)=> (
            <li key={place.xid}>{place.name} — {place.latitude}, {place.longitude}</li>
          ))}
        </ul>
      ) : <div>No items</div>}
    </div>
  )
}
