import { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import { resolveUrl } from '../utils';
import { Card } from 'primereact/card';
import { Button } from 'primereact/button';
import { Dropdown } from 'primereact/dropdown';
import { ProgressSpinner } from 'primereact/progressspinner';
import ItemsDialog from './ItemsDialog';

const API_BASE = 'http://localhost:8081/api/v1';

function FriendWishlists({ toast }) {
  const [friends, setFriends] = useState([]);
  const [selectedFriend, setSelectedFriend] = useState(null);
  const [wishlists, setWishlists] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedWishlist, setSelectedWishlist] = useState(null);

  useEffect(() => {
    axios.get(`${API_BASE}/friends`)
      .then(r => setFriends(r.data.friends || []))
      .catch(() => {});
  }, []);

  const loadFriendWishlists = useCallback(async (friendId) => {
    setLoading(true);
    try {
      const r = await axios.get(`${API_BASE}/wishlists?user_id=${friendId}&page=1&per_page=50`);
      setWishlists(r.data.items || []);
    } catch {
      toast.current?.show({ severity: 'error', summary: 'шибка загрузки', life: 2000 });
    } finally {
      setLoading(false);
    }
  }, [toast]);

  const handleFriendSelect = (friend) => {
    setSelectedFriend(friend);
    loadFriendWishlists(friend.id);
  };

  const friendOptions = friends.map(f => ({ label: f.username, value: f }));

  return (
    <div>
      <div className="mb-4">
        <Dropdown
          value={selectedFriend}
          options={friendOptions}
          onChange={e => handleFriendSelect(e.value)}
          placeholder="ыберите друга"
          className="w-full md:w-20rem"
        />
      </div>

      {loading ? (
        <div className="flex justify-content-center p-4">
          <ProgressSpinner />
        </div>
      ) : selectedFriend && wishlists.length === 0 ? (
        <p className="text-color-secondary text-center"> {selectedFriend.username} нет публичных вишлистов.</p>
      ) : (
        <div className="grid">
          {wishlists.map(w => (
            <div key={w.id} className="col-12 md:col-6 lg:col-4">
              <Card
                title={w.title}
                subTitle={w.description}
                className="h-full"
                footer={
                  <Button
                    label="осмотреть"
                    onClick={() => setSelectedWishlist(w)}
                    className="w-full"
                  />
                }
              >
                {w.imageUrl && (
                  <img
                    src={resolveUrl(w.imageUrl)}
                    alt={w.title}
                    style={{ width: '100%', maxHeight: '160px', objectFit: 'cover', borderRadius: '4px' }}
                    className="mb-3"
                  />
                )}
                {w.eventName && (
                  <div className="flex align-items-center gap-2 text-sm">
                    <i className="pi pi-calendar"></i>
                    <span>{w.eventName}</span>
                    {w.eventDate && (
                      <span className="text-color-secondary">
                        ({new Date(w.eventDate).toLocaleDateString('ru-RU')})
                      </span>
                    )}
                  </div>
                )}
              </Card>
            </div>
          ))}
        </div>
      )}

      {selectedWishlist && (
        <ItemsDialog
          wishlist={selectedWishlist}
          visible={!!selectedWishlist}
          onHide={() => setSelectedWishlist(null)}
          toast={toast}
        />
      )}
    </div>
  );
}

export default FriendWishlists;